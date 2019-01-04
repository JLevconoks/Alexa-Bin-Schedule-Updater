#!/usr/bin/env bash

echo "Building Alexa Bin Schedule Updater"

env GOOS=linux GOARCH=amd64 go build .

echo "Preparing zip package"
zip -j alexa-bin-schedule-updater.zip alexa-bin-schedule-updater

echo "Deploying..."
aws lambda update-function-code --function-name AlexaBinScheduleUpdater --zip-file fileb://alexa-bin-schedule-updater.zip

echo "Cleanup"
rm alexa-bin-schedule-updater
rm alexa-bin-schedule-updater.zip

echo "Done!"