#!/bin/bash

# Set the target organization and space
cf target -o nr-firehose-nozzle-org -s nr-firehose-nozzle-space

# Get the CF apps output
APPS_OUTPUT=$(cf apps)

# Extract the URL from the CF apps output
URL=$(echo "$APPS_OUTPUT" | grep 'newrelic-firehose-nozzle' | awk '{print $6}')

# Perform a health check using the extracted URL
if [ -n "$URL" ]; then
  curl http://$URL/health
else
  echo "Unable to find URL for the newrelic-firehose-nozzle app."
fi
