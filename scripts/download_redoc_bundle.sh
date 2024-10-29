#!/usr/bin/env sh

SCRIPT_PATH=./api/redoc.standalone.js
REDOC_URL=https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js

if [[ ! -f $SCRIPT_PATH || $1 == "update" ]]; then
  echo "Downloading Redoc JavaScript bundle..."
  curl -o $SCRIPT_PATH $REDOC_URL
else
  echo "Redoc JavaScript bundle already downloaded, skipping download"
fi
