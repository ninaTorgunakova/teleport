// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dynamomigration

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"

	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/lib/events"
	"github.com/gravitational/teleport/lib/events/athena"
)

type Config struct {
	// ExportTime is time in the past from which to export table data.
	ExportTime time.Time
	// FreshnessWindow defines if already ongoing or completed exports can be used.
	// For example if freshness is set to 24h any ongoing/completed export from ExportTime-24h will be used
	// instead of triggering new.
	// If no set, new export is triggered.
	FreshnessWindow time.Duration

	// DynamoTableARN that will be exported.
	DynamoTableARN string

	// Bucket used to store export.
	Bucket string
	// Prefix is s3 prefix where to store export inside bucket.
	Prefix string

	// DryRun allows to generate export and convert it to AuditEvents.
	// Nothing is published to athena publisher.
	// Can be used to test if export is valid.
	DryRun bool

	// NoOfEmitWorkers defines how many workers are used to emit audit events.
	NoOfEmitWorkers int
	bufferSize      int

	// TopicARN is topic of athena logger.
	TopicARN string
	// LargePayloadBucket is s3 bucket configured for large payloads in athena logger.
	LargePayloadBucket string
	// LargePayloadPrefix is s3 prefix configured for large payloads in athena logger.
	LargePayloadPrefix string

	Logger log.FieldLogger
}

func (cfg *Config) CheckAndSetDefaults() error {
	if cfg.ExportTime.IsZero() {
		cfg.ExportTime = time.Now()
	}
	if cfg.DynamoTableARN == "" {
		return trace.BadParameter("missing dynamo table ARN")
	}
	if cfg.Bucket == "" {
		return trace.BadParameter("missing export bucket")
	}
	if cfg.NoOfEmitWorkers == 0 {
		cfg.NoOfEmitWorkers = 3
	}
	if cfg.bufferSize == 0 {
		cfg.bufferSize = 10 * cfg.NoOfEmitWorkers
	}
	if !cfg.DryRun {
		if cfg.TopicARN == "" {
			return trace.BadParameter("missing Athena logger SNS Topic ARN")
		}

		if cfg.LargePayloadBucket == "" {
			return trace.BadParameter("missing Athena logger large payload bucket")
		}
	}
	if cfg.Logger == nil {
		cfg.Logger = log.New()
	}
	return nil
}

type task struct {
	Config
	dynamoClient  *dynamodb.Client
	s3Downloader  s3downloader
	eventsEmitter eventsEmitter
}

type s3downloader interface {
	Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*manager.Downloader)) (n int64, err error)
}

type eventsEmitter interface {
	EmitAuditEvent(ctx context.Context, in apievents.AuditEvent) error
}

func newMigrateTask(ctx context.Context, cfg Config) (*task, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	s3Client := s3.NewFromConfig(awsCfg)
	return &task{
		Config:       cfg,
		dynamoClient: dynamodb.NewFromConfig(awsCfg),
		s3Downloader: manager.NewDownloader(s3Client),
		eventsEmitter: athena.NewPublisher(athena.PublisherConfig{
			TopicARN: cfg.TopicARN,
			SNSPublisher: sns.NewFromConfig(awsCfg, func(o *sns.Options) {
				o.Retryer = retry.NewStandard(func(so *retry.StandardOptions) {
					so.MaxAttempts = 30
					so.MaxBackoff = 1 * time.Minute
					// Use bigger rate limit to handle default sdk throttling: https://github.com/aws/aws-sdk-go-v2/issues/1665
					so.RateLimiter = ratelimit.NewTokenRateLimit(1000000)
				})
			}),
			Uploader:      manager.NewUploader(s3Client),
			PayloadBucket: cfg.LargePayloadBucket,
			PayloadPrefix: cfg.LargePayloadPrefix,
		}),
	}, nil
}

// Migrate executed dynamodb -> athena migration.
func Migrate(ctx context.Context, cfg Config) error {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	t, err := newMigrateTask(ctx, cfg)
	if err != nil {
		return trace.Wrap(err)
	}

	dataObjectsInfo, err := t.GetOrStartExportAndWaitForResults(ctx)
	if err != nil {
		return trace.Wrap(err)
	}

	if err := t.ProcessDataObjects(ctx, dataObjectsInfo); err != nil {
		return trace.Wrap(err)
	}

	t.Logger.Info("Migration finished")
	return nil
}

// GetOrStartExportAndWaitForResults return export results.
// It can either reused existing export or start new one, depending on FreshnessWindow.
func (t *task) GetOrStartExportAndWaitForResults(ctx context.Context) ([]dataObjectInfo, error) {
	var exportArn string
	var err error
	if t.FreshnessWindow > 0 {
		exportArn, err = t.isOngoingOrCompletedExport(ctx)
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}
	if exportArn == "" {
		exportArn, err = t.startExportJob(ctx)
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}

	manifest, err := t.waitForCompletedExport(ctx, exportArn)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	log.Debugf("Using export manifest %s", manifest)

	dataObjectsInfo, err := t.getDataObjectsInfo(ctx, manifest)
	return dataObjectsInfo, trace.Wrap(err)
}

// ProcessDataObjects takes dataObjectInfo from export summary, downloads data files
// from s3, ungzip them and emitt them on SNS using athena publisher.
func (t *task) ProcessDataObjects(ctx context.Context, dataObjectsInfo []dataObjectInfo) error {
	eventsC := make(chan apievents.AuditEvent, t.bufferSize)
	done := make(chan struct{})
	go func() {
		t.emitEvents(ctx, eventsC)
		close(done)
	}()

	if err := t.getEventsFromDataFiles(ctx, dataObjectsInfo, eventsC); err != nil {
		return trace.Wrap(err)
	}

	close(eventsC)
	<-done
	return nil
}

func (t *task) isOngoingOrCompletedExport(ctx context.Context) (string, error) {
	// Let's assume that there are not a lot exports and ignore paging at list endpoint.
	exports, err := t.dynamoClient.ListExports(ctx, &dynamodb.ListExportsInput{
		MaxResults: aws.Int32(25),
		TableArn:   aws.String(t.DynamoTableARN),
	})
	if err != nil {
		return "", err
	}
	for _, e := range exports.ExportSummaries {
		switch e.ExportStatus {
		case dynamoTypes.ExportStatusFailed:
			continue
		case dynamoTypes.ExportStatusCompleted, dynamoTypes.ExportStatusInProgress:
			// check if export is from date that we are intrested.
			desc, err := t.dynamoClient.DescribeExport(ctx, &dynamodb.DescribeExportInput{
				ExportArn: e.ExportArn,
			})
			if err != nil {
				continue
			}
			if desc.ExportDescription.ExportTime.Sub(t.ExportTime).Abs() > t.FreshnessWindow {
				// too old, ignore it.
				continue
			}
			// fresh export, return it.
			t.Logger.Infof("Found fresh export from %v, will use it", desc.ExportDescription.ExportTime)
			return aws.ToString(e.ExportArn), nil
		}
	}
	return "", nil
}

func (t *task) waitForCompletedExport(ctx context.Context, exportARN string) (exportManifest string, err error) {
	for {
		exportStatusOutput, err := t.dynamoClient.DescribeExport(ctx, &dynamodb.DescribeExportInput{
			ExportArn: aws.String(exportARN),
		})
		if err != nil {
			return "", trace.Wrap(err)
		}

		if exportStatusOutput.ExportDescription.ExportStatus == dynamoTypes.ExportStatusCompleted {
			return *exportStatusOutput.ExportDescription.ExportManifest, nil
		}

		select {
		case <-ctx.Done():
			return "", trace.Wrap(ctx.Err())
		case <-time.After(10 * time.Second):
			t.Logger.Debugf("Export job still in progress...")
		}
	}
}

func (t *task) startExportJob(ctx context.Context) (arn string, err error) {
	exportOutput, err := t.dynamoClient.ExportTableToPointInTime(ctx, &dynamodb.ExportTableToPointInTimeInput{
		S3Bucket: aws.String(t.Bucket),
		TableArn: aws.String(t.DynamoTableARN),
		// ClientToken:  aws.String(fmt.Sprintf("export-%s", ExportTime.Format(time.DateOnly))),
		ExportFormat: dynamoTypes.ExportFormatDynamodbJson,
		ExportTime:   aws.Time(t.ExportTime),
		S3Prefix:     aws.String(t.Prefix),
	})
	if err != nil {
		return "", trace.Wrap(err)
	}
	t.Logger.Infof("Started new export, it can take a while")
	return aws.ToString(exportOutput.ExportDescription.ExportArn), nil
}

type dataObjectInfo struct {
	DataFileS3Key string `json:"dataFileS3Key"`
	ItemCount     int    `json:"itemCount"`
}

// getDataObjectsInfo downloads manifest-files.json and get data object info from it.
func (t *task) getDataObjectsInfo(ctx context.Context, manifestPath string) ([]dataObjectInfo, error) {
	// summary file is small, we can use in-memory buffer.
	writeAtBuf := manager.NewWriteAtBuffer([]byte{})
	if _, err := t.s3Downloader.Download(ctx, writeAtBuf, &s3.GetObjectInput{
		Bucket: aws.String(t.Bucket),
		// AWS SDK returns manifest-summary.json path. We are interested in
		// manifest-files.json because it's contains references about data export files.
		Key: aws.String(path.Dir(manifestPath) + "/manifest-files.json"),
	}); err != nil {
		return nil, trace.Wrap(err)
	}

	out := make([]dataObjectInfo, 0)
	scanner := bufio.NewScanner(bytes.NewBuffer(writeAtBuf.Bytes()))
	// manifest-files are JSON lines files, that why it's scanned line by line.
	for scanner.Scan() {
		var obj dataObjectInfo
		err := json.Unmarshal(scanner.Bytes(), &obj)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		out = append(out, obj)
	}
	if err := scanner.Err(); err != nil {
		return nil, trace.Wrap(err)
	}
	return out, nil
}

func (t *task) getEventsFromDataFiles(ctx context.Context, dataObjectsInfo []dataObjectInfo, eventsC chan<- apievents.AuditEvent) error {
	for _, dataObj := range dataObjectsInfo {
		t.Logger.Debugf("Downloading %s", dataObj.DataFileS3Key)

		if err := t.fromS3ToChan(ctx, dataObj, eventsC); err != nil {
			return trace.Wrap(err)
		}
	}
	return nil
}

func (t *task) fromS3ToChan(ctx context.Context, dataObj dataObjectInfo, eventsC chan<- apievents.AuditEvent) error {
	f, err := t.downloadFromS3(ctx, dataObj.DataFileS3Key)
	if err != nil {
		return trace.Wrap(err)
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return trace.Wrap(err)
	}
	defer gzipReader.Close()

	scanner := bufio.NewScanner(gzipReader)
	t.Logger.Debugf("Scanning %d events", dataObj.ItemCount)
	count := 0
	for scanner.Scan() {
		count++
		ev, err := exportedDynamoItemToAuditEvent(ctx, scanner.Bytes())
		if err != nil {
			return trace.Wrap(err)
		}
		select {
		case eventsC <- ev:
		case <-ctx.Done():
			return ctx.Err()
		}

		if count%100 == 0 {
			t.Logger.Debugf("Sent on buffer %d/%d events", count, dataObj.ItemCount)
		}
	}

	if err := scanner.Err(); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// exportedDynamoItemToAuditEvent converts single line of dynamo export into AuditEvent.
func exportedDynamoItemToAuditEvent(ctx context.Context, in []byte) (apievents.AuditEvent, error) {
	var itemMap map[string]map[string]any
	if err := json.Unmarshal(in, &itemMap); err != nil {
		return nil, trace.Wrap(err)
	}

	var attributeMap map[string]dynamoTypes.AttributeValue
	if err := awsAwsjson10_deserializeDocumentAttributeMap(&attributeMap, itemMap["Item"]); err != nil {
		return nil, trace.Wrap(err)
	}

	var eventFields events.EventFields
	if err := attributevalue.Unmarshal(attributeMap["FieldsMap"], &eventFields); err != nil {
		return nil, trace.Wrap(err)
	}

	event, err := events.FromEventFields(eventFields)
	return event, trace.Wrap(err)
}

func (t *task) downloadFromS3(ctx context.Context, key string) (*os.File, error) {
	originalName := path.Base(key)
	path := path.Join(os.TempDir(), originalName)

	f, err := os.Create(path)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if _, err := t.s3Downloader.Download(ctx, f, &s3.GetObjectInput{
		Bucket: aws.String(t.Bucket),
		Key:    aws.String(key),
	}); err != nil {
		f.Close()
		return nil, trace.Wrap(err)
	}
	return f, nil
}

func (t *task) emitEvents(ctx context.Context, eventsC <-chan apievents.AuditEvent) {
	if t.DryRun {
		// in dryRun we just want to count events, validation is done on ready from file.
		var count int
		var oldest, newest apievents.AuditEvent
		for event := range eventsC {
			count++
			if oldest == nil && newest == nil {
				// first iteration, initialize values with first event.
				oldest = event
				newest = event
			}
			if oldest.GetTime().After(event.GetTime()) {
				oldest = event
			}
			if newest.GetTime().Before(event.GetTime()) {
				newest = event
			}
		}
		t.Logger.Infof("Dry run: there are %d events from %v to %v", count, oldest.GetTime(), newest.GetTime())
		return
	}

	var wg sync.WaitGroup
	wg.Add(t.NoOfEmitWorkers)
	for i := 0; i < t.NoOfEmitWorkers; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-eventsC:
					if !ok {
						return
					}
					if err := t.eventsEmitter.EmitAuditEvent(ctx, e); err != nil {
						// TODO(tobiaszheller): rethink what we should do in that case.
						// If someone uses dryRun first, it will very that export is valid.
						// If we cannot emmit event, it's probably error with configuration of SNS or rate limit/transient error on AWS.
						t.Logger.Errorf("Failed to publish event %s of type %s: %v", e.GetID(), e.GetType(), err)
					}
				}
			}
		}()
	}
	wg.Wait()
}
