#!/bin/bash
set -eo pipefail

echo "=================================================="
echo "    🧪 NexusLink Load Testing Script (Vegeta)       "
echo "=================================================="

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null
then
    echo "Vegeta is not installed. Please install it first:"
    echo "go install github.com/tsenart/vegeta/v12@latest"
    exit 1
fi

TARGET_URL=${1:-"http://localhost:8080"}
DURATION=${2:-"15s"}
RATE=${3:-"500"}

echo "Running load test..."
echo "Target: $TARGET_URL"
echo "Rate: $RATE req/s"
echo "Duration: $DURATION"
echo "--------------------------------------------------"

echo "GET $TARGET_URL/" | vegeta attack -duration="$DURATION" -rate="$RATE" | vegeta report

echo "=================================================="
echo "Load test completed."
