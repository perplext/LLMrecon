#!/bin/bash
# Script to optimize templates and measure performance improvements

set -e

# Configuration
TEMPLATE_SOURCE="./templates"
SOURCE_TYPE="file"
OUTPUT_DIR="./optimized-templates"
ITERATIONS=5
CONCURRENCY=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
CACHE_SIZE=1000
CACHE_TTL=3600

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
    --output)
      OUTPUT_DIR="$2"
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
    --help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  --source <path>       Template source path (default: ./templates)"
      echo "  --type <type>         Source type (file, github, gitlab) (default: file)"
      echo "  --output <dir>        Output directory for optimized templates (default: ./optimized-templates)"
      echo "  --iterations <num>    Number of benchmark iterations (default: 5)"
      echo "  --concurrency <num>   Concurrency limit (default: number of CPU cores)"
      echo "  --cache-size <num>    Maximum cache size (default: 1000)"
      echo "  --cache-ttl <seconds> Cache TTL in seconds (default: 3600)"
      echo "  --help                Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $key"
      exit 1
      ;;
  esac
done

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build the template optimizer tool
echo "Building template optimizer tool..."
go build -o ./bin/template-optimizer ./cmd/template-optimizer

# Run benchmark with standard loader
echo "Running benchmark with standard loader..."
./bin/template-optimizer \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --benchmark \
  --iterations="$ITERATIONS" \
  --concurrency="$CONCURRENCY" \
  --cache-size="$CACHE_SIZE" \
  --cache-ttl="$CACHE_TTL" \
  --optimize=false \
  --structure-optimize=false \
  --verbose > standard-benchmark.txt

# Run benchmark with optimized loader
echo "Running benchmark with optimized loader..."
./bin/template-optimizer \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --benchmark \
  --iterations="$ITERATIONS" \
  --concurrency="$CONCURRENCY" \
  --cache-size="$CACHE_SIZE" \
  --cache-ttl="$CACHE_TTL" \
  --optimize=true \
  --structure-optimize=true \
  --verbose > optimized-benchmark.txt

# Generate optimized templates
echo "Generating optimized templates..."
./bin/template-optimizer \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --output="$OUTPUT_DIR" \
  --concurrency="$CONCURRENCY" \
  --cache-size="$CACHE_SIZE" \
  --cache-ttl="$CACHE_TTL" \
  --optimize=true \
  --structure-optimize=true \
  --verbose

# Compare benchmark results
echo "Benchmark Results Comparison:"
echo "============================"
echo "Standard Loader:"
grep -A 3 "Benchmark Results:" standard-benchmark.txt

echo ""
echo "Optimized Loader:"
grep -A 3 "Benchmark Results:" optimized-benchmark.txt

# Calculate improvement percentage
STD_TIME=$(grep "Average Load Time:" standard-benchmark.txt | awk '{print $4}')
OPT_TIME=$(grep "Average Load Time:" optimized-benchmark.txt | awk '{print $4}')

if [[ "$STD_TIME" =~ ([0-9]+)([a-z]+) ]]; then
  STD_VALUE=${BASH_REMATCH[1]}
  STD_UNIT=${BASH_REMATCH[2]}
  
  if [[ "$OPT_TIME" =~ ([0-9]+)([a-z]+) ]]; then
    OPT_VALUE=${BASH_REMATCH[1]}
    OPT_UNIT=${BASH_REMATCH[2]}
    
    # Convert to same unit if needed
    if [[ "$STD_UNIT" == "$OPT_UNIT" ]]; then
      IMPROVEMENT=$(echo "scale=2; (($STD_VALUE - $OPT_VALUE) / $STD_VALUE) * 100" | bc)
      echo ""
      echo "Performance Improvement: $IMPROVEMENT%"
    else
      echo ""
      echo "Units differ, cannot calculate improvement percentage"
    fi
  fi
fi

echo ""
echo "Optimized templates saved to: $OUTPUT_DIR"
echo "Benchmark reports saved to: standard-benchmark.txt and optimized-benchmark.txt"
