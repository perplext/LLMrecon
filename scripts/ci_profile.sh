#!/bin/bash
# CI/CD script for template performance profiling and baseline comparison

set -e

# Configuration
TEMPLATE_SOURCE="./templates"
SOURCE_TYPE="file"
REPORT_DIR="./performance-reports"
BASELINE_FILE="./performance-reports/baseline.json"
THRESHOLD_FILE="./performance-reports/thresholds.json"
ITERATIONS=5
CONCURRENCY=10
CACHE_SIZE=1000
CACHE_TTL=3600
FAIL_ON_THRESHOLD=true

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --source)
      TEMPLATE_SOURCE="$2"
      shift
      shift
      ;;
    --type)
      SOURCE_TYPE="$2"
      shift
      shift
      ;;
    --report-dir)
      REPORT_DIR="$2"
      shift
      shift
      ;;
    --baseline-file)
      BASELINE_FILE="$2"
      shift
      shift
      ;;
    --threshold-file)
      THRESHOLD_FILE="$2"
      shift
      shift
      ;;
    --iterations)
      ITERATIONS="$2"
      shift
      shift
      ;;
    --concurrency)
      CONCURRENCY="$2"
      shift
      shift
      ;;
    --cache-size)
      CACHE_SIZE="$2"
      shift
      shift
      ;;
    --cache-ttl)
      CACHE_TTL="$2"
      shift
      shift
      ;;
    --establish-baseline)
      ESTABLISH_BASELINE=true
      shift
      ;;
    --compare)
      COMPARE=true
      shift
      ;;
    --no-fail)
      FAIL_ON_THRESHOLD=false
      shift
      ;;
    *)
      echo "Unknown option: $key"
      exit 1
      ;;
  esac
done

# Create report directory if it doesn't exist
mkdir -p "$REPORT_DIR"

# Build the profiling tool if needed
echo "Building profiling tool..."
go build -o ./bin/profile ./cmd/profile

# Establish baseline if requested
if [ "$ESTABLISH_BASELINE" = true ]; then
  echo "Establishing baseline metrics..."
  ./bin/profile \
    --source="$TEMPLATE_SOURCE" \
    --type="$SOURCE_TYPE" \
    --iterations="$ITERATIONS" \
    --concurrency="$CONCURRENCY" \
    --cache-size="$CACHE_SIZE" \
    --cache-ttl="$CACHE_TTL" \
    --baseline \
    --baseline-file="$BASELINE_FILE" \
    --report="$REPORT_DIR/baseline-report.txt" \
    --optimized=true \
    --verbose=true
  
  echo "Baseline metrics established and saved to $BASELINE_FILE"
  exit 0
fi

# Run profiling
echo "Running performance profiling..."
./bin/profile \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --iterations="$ITERATIONS" \
  --concurrency="$CONCURRENCY" \
  --cache-size="$CACHE_SIZE" \
  --cache-ttl="$CACHE_TTL" \
  --report="$REPORT_DIR/profile-report.txt" \
  --optimized=true \
  --verbose=true

# Compare with baseline if requested
if [ "$COMPARE" = true ]; then
  echo "Comparing with baseline metrics..."
  ./bin/profile \
    --source="$TEMPLATE_SOURCE" \
    --type="$SOURCE_TYPE" \
    --compare \
    --baseline-file="$BASELINE_FILE" \
    --report="$REPORT_DIR/comparison-report.txt" \
    --optimized=true
  
  # Check if thresholds were exceeded
  if grep -q "EXCEEDED" "$REPORT_DIR/comparison-report.txt"; then
    echo "⚠️ Performance thresholds exceeded!"
    
    if [ "$FAIL_ON_THRESHOLD" = true ]; then
      echo "❌ CI/CD pipeline failed due to exceeded performance thresholds"
      exit 1
    else
      echo "⚠️ Continuing despite exceeded thresholds (--no-fail option enabled)"
    fi
  else
    echo "✅ All performance metrics are within acceptable thresholds"
  fi
fi

echo "Performance profiling completed successfully"
exit 0
