#!/bin/bash
# Script to benchmark different template execution strategies

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

# Build the concurrent executor tool
echo "Building concurrent executor tool..."
go build -o ./bin/concurrent-executor ./cmd/concurrent-executor

# Run benchmark with standard executor
echo "Running benchmark with standard executor..."
./bin/concurrent-executor \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --executor="standard" \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > standard-executor.txt

# Run benchmark with optimized executor
echo "Running benchmark with optimized executor..."
./bin/concurrent-executor \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --executor="optimized" \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > optimized-executor.txt

# Run benchmark with async executor
echo "Running benchmark with async executor..."
./bin/concurrent-executor \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --executor="async" \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > async-executor.txt

# Run benchmark with pipeline executor
echo "Running benchmark with pipeline executor..."
./bin/concurrent-executor \
  --source="$TEMPLATE_SOURCE" \
  --type="$SOURCE_TYPE" \
  --executor="pipeline" \
  --iterations="$ITERATIONS" \
  --count="$TEMPLATE_COUNT" \
  --response-time="$RESPONSE_TIME" \
  --concurrency="$CONCURRENCY" > pipeline-executor.txt

# Compare benchmark results
echo "Benchmark Results Comparison:"
echo "============================"
echo "Standard Executor:"
grep -A 3 "Benchmark Results:" standard-executor.txt

echo ""
echo "Optimized Executor:"
grep -A 3 "Benchmark Results:" optimized-executor.txt

echo ""
echo "Async Executor:"
grep -A 3 "Benchmark Results:" async-executor.txt

echo ""
echo "Pipeline Executor:"
grep -A 3 "Benchmark Results:" pipeline-executor.txt

# Calculate improvement percentages
STD_TIME=$(grep "Average Execution Time:" standard-executor.txt | awk '{print $4}')
OPT_TIME=$(grep "Average Execution Time:" optimized-executor.txt | awk '{print $4}')
ASYNC_TIME=$(grep "Average Execution Time:" async-executor.txt | awk '{print $4}')
PIPE_TIME=$(grep "Average Execution Time:" pipeline-executor.txt | awk '{print $4}')

if [[ "$STD_TIME" =~ ([0-9]+)([a-z]+) ]]; then
  STD_VALUE=${BASH_REMATCH[1]}
  STD_UNIT=${BASH_REMATCH[2]}
  
  # Calculate optimized improvement
  if [[ "$OPT_TIME" =~ ([0-9]+)([a-z]+) ]] && [[ "${BASH_REMATCH[2]}" == "$STD_UNIT" ]]; then
    OPT_VALUE=${BASH_REMATCH[1]}
    OPT_IMPROVEMENT=$(echo "scale=2; (($STD_VALUE - $OPT_VALUE) / $STD_VALUE) * 100" | bc)
    echo ""
    echo "Optimized Executor Improvement: $OPT_IMPROVEMENT%"
  fi
  
  # Calculate async improvement
  if [[ "$ASYNC_TIME" =~ ([0-9]+)([a-z]+) ]] && [[ "${BASH_REMATCH[2]}" == "$STD_UNIT" ]]; then
    ASYNC_VALUE=${BASH_REMATCH[1]}
    ASYNC_IMPROVEMENT=$(echo "scale=2; (($STD_VALUE - $ASYNC_VALUE) / $STD_VALUE) * 100" | bc)
    echo "Async Executor Improvement: $ASYNC_IMPROVEMENT%"
  fi
  
  # Calculate pipeline improvement
  if [[ "$PIPE_TIME" =~ ([0-9]+)([a-z]+) ]] && [[ "${BASH_REMATCH[2]}" == "$STD_UNIT" ]]; then
    PIPE_VALUE=${BASH_REMATCH[1]}
    PIPE_IMPROVEMENT=$(echo "scale=2; (($STD_VALUE - $PIPE_VALUE) / $STD_VALUE) * 100" | bc)
    echo "Pipeline Executor Improvement: $PIPE_IMPROVEMENT%"
  fi
fi

echo ""
echo "Benchmark reports saved to: standard-executor.txt, optimized-executor.txt, async-executor.txt, and pipeline-executor.txt"
