# Dynamomigration tool

Dynamomigration tool allows to export Teleport audit events logs from DynamoDB
table into Athena Audit log.
It's using DynamoDB export to S3 to export data.

Requirements:

* Point-in-time recovery (PITR) on DynamoDB table
* IAM permiossions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowDynamoExportAndList",
            "Effect": "Allow",
            "Action": [
                "dynamodb:ExportTableToPointInTime"
            ],
            "Resource": "arn:aws:dynamodb:region:account:table/tablename"
        },
        {
            "Sid": "AllowDynamoExportDescribe",
            "Effect": "Allow",
            "Action": [
                "dynamodb:DescribeExport"
            ],
            "Resource": "arn:aws:dynamodb:region:account:table/tablename/*"
        },
        {
            "Sid": "AllowWriteReadDestinationBucket",
            "Effect": "Allow",
            "Action": [
                "s3:AbortMultipartUpload",
                "s3:PutObject",
                "s3:PutObjectAcl",
                "s3:GetObject"
            ],
            "Resource": "arn:aws:s3:::export-bucket/*"
        },
        {
            "Sid": "AllowWriteLargePayloadsBucket",
            "Effect": "Allow",
            "Action": [
                "s3:AbortMultipartUpload",
                "s3:PutObject",
                "s3:PutObjectAcl"
            ],
            "Resource": "arn:aws:s3:::large-payloads-bucket/*"
        },
        {
            "Sid": "AllowPublishToAthenaTopic",
            "Effect": "Allow",
            "Action": [
                "sns:Publish"
            ],
            "Resource": "arn:aws:sns:region:account:topicname"
        }
    ]
}
```

## Example usage

Build: `cd lib/events/athena/dynamomigration/cmd && go build -o dynamomigration`.

It is recommended to test export first using `-dryRun` flag. DryRun does not emit any events,
it makes sure that export is in valid format and events can be parsed.

Dry run example:

```shell
./dynamomigration -dynamoARN='arn:aws:dynamodb:region:account:table/tablename' \
  -exportPath='s3://bucket/prefix' \
  -freshnessWindow=24h \
  -dryRun
```

Full migration:

```shell
./dynamomigration -dynamoARN='arn:aws:dynamodb:region:account:table/tablename' \
  -exportPath='s3://bucket/prefix' \
  -freshnessWindow=24h \
  -snsTopicARN=arn:aws:sns:region:account:topicname \
  -largePayloadsPath=s3://bucket/prefix
```
