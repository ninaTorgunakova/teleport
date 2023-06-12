/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/retryutils"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
)

// minBatchSize is the minimum batch size to send ServerInfos in for discovered
// instances.
const minBatchSize = 5

type labelReconcilerConfig struct {
	clock       clockwork.Clock
	log         logrus.FieldLogger
	accessPoint auth.DiscoveryAccessPoint
}

func (c *labelReconcilerConfig) checkAndSetDefaults() error {
	if c.accessPoint == nil {
		return trace.BadParameter("missing parameter: accessPoint")
	}
	if c.clock == nil {
		c.clock = clockwork.NewRealClock()
	}
	if c.log == nil {
		c.log = logrus.New()
	}
	return nil
}

type labelReconciler struct {
	cfg *labelReconcilerConfig

	mu                sync.Mutex
	discoveredServers map[string]types.ServerInfo
	serverInfoQueue   []types.ServerInfo
	lastBatchSize     int
}

func newLabelReconciler(cfg *labelReconcilerConfig) (*labelReconciler, error) {
	if err := cfg.checkAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return &labelReconciler{
		cfg:               cfg,
		discoveredServers: make(map[string]types.ServerInfo),
		serverInfoQueue:   make([]types.ServerInfo, 0, minBatchSize),
		lastBatchSize:     minBatchSize,
	}, nil
}

// getUpsertBatchSize calculates the size of batch to upsert ServerInfos in.
//
// Batches are sent once per second, and the goal is to upsert all ServerInfos
// within 15 minutes.
func getUpsertBatchSize(queueLen, lastBatchSize int) int {
	batchSize := lastBatchSize
	// Increase batch size so that all upserts can finish within 15 minutes.
	if dynamicBatchSize := (queueLen / 900) + 1; dynamicBatchSize > batchSize {
		batchSize = dynamicBatchSize
	}
	if batchSize < minBatchSize {
		batchSize = minBatchSize
	}
	if batchSize > queueLen {
		batchSize = queueLen
	}
	return batchSize
}

func (r *labelReconciler) run(ctx context.Context) {
	ticker := r.cfg.clock.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.Chan():
			r.mu.Lock()
			if len(r.serverInfoQueue) == 0 {
				r.mu.Unlock()
				continue
			}

			batchSize := getUpsertBatchSize(len(r.serverInfoQueue), r.lastBatchSize)
			r.lastBatchSize = batchSize
			batch := r.serverInfoQueue[:batchSize]
			r.serverInfoQueue = r.serverInfoQueue[batchSize:]
			r.mu.Unlock()

			for _, si := range batch {
				if err := r.cfg.accessPoint.UpsertServerInfo(ctx, si); err != nil {
					r.cfg.log.WithError(err).Error("Failed to upsert server info.")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// queueServerInfos queues a list of ServerInfos to be upserted.
func (r *labelReconciler) queueServerInfos(serverInfos []types.ServerInfo) {
	jitter := retryutils.NewSeventhJitter()
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.cfg.clock.Now()
	for _, si := range serverInfos {
		si.SetExpiry(now.Add(jitter(90 * time.Minute)))
		existingInfo, ok := r.discoveredServers[si.GetName()]
		// ServerInfos should be upserted if
		//   - the instance is new
		//   - the instance's labels have changed
		//   - the existing ServerInfo will expire within 30 minutes
		if !ok ||
			!utils.StringMapsEqual(si.GetStaticLabels(), existingInfo.GetStaticLabels()) ||
			existingInfo.Expiry().Before(now.Add(30*time.Minute)) {
			r.discoveredServers[si.GetName()] = si
			r.serverInfoQueue = append(r.serverInfoQueue, si)
		}
	}
}
