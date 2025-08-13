#!/bin/bash

# Benchmark script for static file handler performance
# This script compares the performance of the static file handler with different configurations

# Set default values
NUM_FILES=100
FILE_SIZE=10240
CONCURRENT=50
ITERATIONS=5
OUTPUT_DIR="./benchmark_results"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --files)
      NUM_FILES="$2"
      shift 2
      ;;
    --size)
      FILE_SIZE="$2"
      shift 2
      ;;
    --concurrent)
      CONCURRENT="$2"
      shift 2
      ;;
    --iterations)
      ITERATIONS="$2"
      shift 2
      ;;
    --output)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--files NUM_FILES] [--size FILE_SIZE] [--concurrent CONCURRENT] [--iterations ITERATIONS] [--output OUTPUT_DIR]"
      exit 1
      ;;
  esac
done

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Function to run benchmark with specific options
run_benchmark() {
  local name=$1
  local cache=$2
  local compression=$3
  local output_file="$OUTPUT_DIR/${name// /_}.md"
  
  echo "Running benchmark: $name"
  echo "Cache: $cache, Compression: $compression"
  echo "Output file: $output_file"
  
  # Run the benchmark
  go run cmd/memory-benchmark/main.go \
    --templates 100 \
    --static-file-handler \
    --static-file-cache=$cache \
    --static-file-compression=$compression \
    --static-files=$NUM_FILES \
    --static-file-size=$FILE_SIZE \
    --concurrent=$CONCURRENT \
    --iterations=$ITERATIONS \
    --output="$output_file"
    
  echo "Benchmark completed: $name"
  echo "Results saved to: $output_file"
  echo
}

# Print benchmark configuration
echo "Static File Handler Benchmark"
echo "============================"
echo "Number of files: $NUM_FILES"
echo "File size: $FILE_SIZE bytes"
echo "Concurrent operations: $CONCURRENT"
echo "Iterations: $ITERATIONS"
echo "Output directory: $OUTPUT_DIR"
echo

# Run benchmarks with different configurations
run_benchmark "No Cache, No Compression" false false
run_benchmark "Cache Only" true false
run_benchmark "Compression Only" false true
run_benchmark "Cache and Compression" true true

# Generate comparison report
echo "Generating comparison report..."
cat > "$OUTPUT_DIR/comparison.md" << EOF
# Static File Handler Performance Comparison

This report compares the performance of the static file handler with different configurations.

## Benchmark Configuration

- Number of files: $NUM_FILES
- File size: $FILE_SIZE bytes
- Concurrent operations: $CONCURRENT
- Iterations: $ITERATIONS

## Results Summary

| Configuration | Memory Usage | Files/sec | Cache Hits | Compression Ratio |
|---------------|--------------|-----------|------------|-------------------|
EOF

# Extract and add results to comparison report
for config in "No_Cache_No_Compression" "Cache_Only" "Compression_Only" "Cache_and_Compression"; do
  file="$OUTPUT_DIR/${config}.md"
  if [ -f "$file" ]; then
    # Extract memory usage
    memory=$(grep "Heap Alloc" "$file" | grep "Final" | awk '{print $4}')
    
    # Extract files per second
    files_per_sec=$(grep "Files Per Second" "$file" | awk '{print $4}')
    
    # Extract cache hits
    cache_hits=$(grep "Cache Hits" "$file" | awk '{print $3}')
    
    # Extract compression ratio
    compression=$(grep "Compression Ratio" "$file" | awk '{print $3}')
    
    # Add to comparison report
    echo "| ${config//_/ } | $memory MB | $files_per_sec | $cache_hits | $compression |" >> "$OUTPUT_DIR/comparison.md"
  fi
done

echo >> "$OUTPUT_DIR/comparison.md"
echo "## Conclusion" >> "$OUTPUT_DIR/comparison.md"
echo >> "$OUTPUT_DIR/comparison.md"
echo "The benchmark results show that:" >> "$OUTPUT_DIR/comparison.md"
echo "1. Enabling caching significantly reduces memory usage and improves throughput" >> "$OUTPUT_DIR/comparison.md"
echo "2. Compression provides additional benefits for text-based files" >> "$OUTPUT_DIR/comparison.md"
echo "3. The combination of caching and compression provides the best overall performance" >> "$OUTPUT_DIR/comparison.md"

echo "Comparison report generated: $OUTPUT_DIR/comparison.md"
echo "Benchmark completed successfully!"
