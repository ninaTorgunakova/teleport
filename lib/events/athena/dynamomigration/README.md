# Dynamomigration tool

Dynamomigration tool allows to export Teleport audit events logs from DynamoDB
table into Athena Audit log.
It's using DynamoDB export to S3 to export data.

Requirements:

* Point-in-time recovery (PITR) on DynamoDB table
* IAM permiossions
TODO(tobiaszheller): add required permissions to run migration.

## Example usage

Build: `cd lib/events/athena/dynamomigration/cmd && go build -o dynamomigration`.

Dry run to test export:

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
