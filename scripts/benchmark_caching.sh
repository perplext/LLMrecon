#!/bin/bash
# Script to benchmark different caching strategies

set -e

# Configuration
TEMPLATE_SOURCE="./templates"
SOURCE_TYPE="file"
ITERATIONS=3
TEMPLATE_COUNT=20
RESPONSE_TIME=100
CONCURRENCY=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

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
    --iterations)
      ITERATIONS="$2"
      shift
      shift
      ;;
    --count)
      TEMPLATE_COUNT="$2"
      shift
      shift
      ;;
    --response-time)
      RESPONSE_TIME="$2"
      shift
      shift
      ;;
    --concurrency)
      CONCURRENCY="$2"
      shift
      shift
      ;;
    --help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  --source <path>       Template source path (default: ./templates)"
      echo "  --type <type>         Source type (file, github, gitlab) (default: file)"
      echo "  --iterations <num>    Number of benchmark iterations (default: 3)"
      echo "  --count <num>         Number of templates to execute (default: 20)"
      echo "  --response-time <ms>  Simulated LLM response time in milliseconds (default: 100)"
      echo "  --concurrency <num>   Concurrency limit (default: number of CPU cores)"
      echo "  --help                Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $key"
      exit 1
      ;;
  esac
done

# Build the cache benchmark tool
echo "Building cache benchmark tool..."
go build -o ./bin/cache-benchmark ./cmd/cache-benchmark

# Run benchmark with no caching
echo "Running benchmark with no caching..."
./bin/cache-benchmark \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --cache=false \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > no-cache.txt

# Run benchmark with standard caching
echo "Running benchmark with standard caching..."
./bin/cache-benchmark \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --cache=true \
  --multi-level=false \
  --repo-cache=false \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > standard-cache.txt

# Run benchmark with multi-level caching
echo "Running benchmark with multi-level caching..."
./bin/cache-benchmark \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --cache=true \
  --multi-level=true \
  --repo-cache=false \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > multi-level-cache.txt

# Run benchmark with repository caching
echo "Running benchmark with repository caching..."
./bin/cache-benchmark \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --cache=true \
  --multi-level=true \
  --repo-cache=true \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > full-cache.txt

# Compare benchmark results
echo "Benchmark Results Comparison:"
echo "============================"
echo "No Caching:"
grep -A 3 "Benchmark Results:" no-cache.txt

echo ""
echo "Standard Caching:"
grep -A 3 "Benchmark Results:" standard-cache.txt

echo ""
echo "Multi-Level Caching:"
grep -A 3 "Benchmark Results:" multi-level-cache.txt

echo ""
echo "Full Caching (Multi-Level + Repository):"
grep -A 3 "Benchmark Results:" full-cache.txt

# Calculate improvement percentages
NO_TIME=$(grep "Average Execution Time:" no-cache.txt | awk '{print $4}')
STD_TIME=$(grep "Average Execution Time:" standard-cache.txt | awk '{print $4}')
MULTI_TIME=$(grep "Average Execution Time:" multi-level-cache.txt | awk '{print $4}')
FULL_TIME=$(grep "Average Execution Time:" full-cache.txt | awk '{print $4}')

if [[ "$NO_TIME" =~ ([0-9]+)([a-z]+) ]]; then
  NO_VALUE=${BASH_REMATCH[1]}
  NO_UNIT=${BASH_REMATCH[2]}
  
  # Calculate standard caching improvement
  if [[ "$STD_TIME" =~ ([0-9]+)([a-z]+) ]] && [[ "${BASH_REMATCH[2]}" == "$NO_UNIT" ]]; then
    STD_VALUE=${BASH_REMATCH[1]}
    STD_IMPROVEMENT=$(echo "scale=2; (($NO_VALUE - $STD_VALUE) / $NO_VALUE) * 100" | bc)
    echo ""
    echo "Standard Caching Improvement: $STD_IMPROVEMENT%"
  fi
  
  # Calculate multi-level caching improvement
  if [[ "$MULTI_TIME" =~ ([0-9]+)([a-z]+) ]] && [[ "${BASH_REMATCH[2]}" == "$NO_UNIT" ]]; then
    MULTI_VALUE=${BASH_REMATCH[1]}
    MULTI_IMPROVEMENT=$(echo "scale=2; (($NO_VALUE - $MULTI_VALUE) / $NO_VALUE) * 100" | bc)
    echo "Multi-Level Caching Improvement: $MULTI_IMPROVEMENT%"
  fi
  
  # Calculate full caching improvement
  if [[ "$FULL_TIME" =~ ([0-9]+)([a-z]+) ]] && [[ "${BASH_REMATCH[2]}" == "$NO_UNIT" ]]; then
    FULL_VALUE=${BASH_REMATCH[1]}
    FULL_IMPROVEMENT=$(echo "scale=2; (($NO_VALUE - $FULL_VALUE) / $NO_VALUE) * 100" | bc)
    echo "Full Caching Improvement: $FULL_IMPROVEMENT%"
  fi
fi

echo ""
echo "Benchmark reports saved to: no-cache.txt, standard-cache.txt, multi-level-cache.txt, and full-cache.txt"
