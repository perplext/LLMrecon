#!/bin/bash

# Memory Optimization and Tuning Script
# This script provides commands to optimize memory usage and tune server parameters

set -e

# Default values
ACTION="status"
MEMORY_THRESHOLD=100
GC_PERCENT=100
MAX_PROCS=0
POOL_SIZE=0
ENABLE_OPTIMIZER=true
ENABLE_POOL_MANAGER=true
ENABLE_TUNER=true
PROFILE_OUTPUT="./profiles"
CONFIG_OUTPUT="./config"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --action)
      ACTION="$2"
      shift 2
      ;;
    --memory-threshold)
      MEMORY_THRESHOLD="$2"
      shift 2
      ;;
    --gc-percent)
      GC_PERCENT="$2"
      shift 2
      ;;
    --max-procs)
      MAX_PROCS="$2"
      shift 2
      ;;
    --pool-size)
      POOL_SIZE="$2"
      shift 2
      ;;
    --enable-optimizer)
      ENABLE_OPTIMIZER="$2"
      shift 2
      ;;
    --enable-pool-manager)
      ENABLE_POOL_MANAGER="$2"
      shift 2
      ;;
    --enable-tuner)
      ENABLE_TUNER="$2"
      shift 2
      ;;
    --profile-output)
      PROFILE_OUTPUT="$2"
      shift 2
      ;;
    --config-output)
      CONFIG_OUTPUT="$2"
      shift 2
      ;;
    --help)
      echo "Memory Optimization and Tuning Script"
      echo ""
      echo "Usage: $0 [options]"
      echo ""
      echo "Options:"
      echo "  --action <action>             Action to perform (status, optimize, tune, profile, benchmark)"
      echo "  --memory-threshold <MB>       Memory threshold for optimization (in MB)"
      echo "  --gc-percent <percent>        Garbage collection target percentage"
      echo "  --max-procs <count>           Maximum number of processors to use (0 = all)"
      echo "  --pool-size <size>            Size of resource pools"
      echo "  --enable-optimizer <bool>     Enable memory optimizer"
      echo "  --enable-pool-manager <bool>  Enable resource pool manager"
      echo "  --enable-tuner <bool>         Enable configuration tuner"
      echo "  --profile-output <dir>        Output directory for profiles"
      echo "  --config-output <dir>         Output directory for configuration"
      echo "  --help                        Show this help message"
      echo ""
      echo "Actions:"
      echo "  status     - Show current memory usage and configuration"
      echo "  optimize   - Run memory optimization"
      echo "  tune       - Run configuration tuning"
      echo "  profile    - Capture memory profile"
      echo "  benchmark  - Run memory benchmark"
      echo ""
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Create output directories
mkdir -p "$PROFILE_OUTPUT"
mkdir -p "$CONFIG_OUTPUT"

# Set GOMAXPROCS if specified
if [ "$MAX_PROCS" -gt 0 ]; then
  export GOMAXPROCS=$MAX_PROCS
  echo "Set GOMAXPROCS to $MAX_PROCS"
fi

# Set GOGC if specified
if [ "$GC_PERCENT" -gt 0 ]; then
  export GOGC=$GC_PERCENT
  echo "Set GOGC to $GC_PERCENT"
fi

# Perform action
case $ACTION in
  status)
    echo "Memory Usage Status"
    echo "------------------"
    
    # Get memory usage with Go tool
    go run ./cmd/memory-benchmark/main.go \
      --templates=1 \
      --size=1 \
      --variables=1 \
      --concurrent=1 \
      --iterations=1 \
      --optimizer=false \
      --poolmanager=false \
      --tuner=false \
      --output=/dev/null \
      --verbose=true
    
    # Show system memory usage
    echo ""
    echo "System Memory Usage"
    echo "------------------"
    if [ "$(uname)" == "Darwin" ]; then
      # macOS
      vm_stat | perl -ne '/page size of (\d+)/ and $size=$1; /Pages free: (\d+)/ and printf "Free Memory: %.2f GB\n", $1 * $size / 1024 / 1024 / 1024; /Pages active: (\d+)/ and printf "Active Memory: %.2f GB\n", $1 * $size / 1024 / 1024 / 1024; /Pages inactive: (\d+)/ and printf "Inactive Memory: %.2f GB\n", $1 * $size / 1024 / 1024 / 1024; /Pages speculative: (\d+)/ and printf "Speculative Memory: %.2f GB\n", $1 * $size / 1024 / 1024 / 1024; /Pages wired down: (\d+)/ and printf "Wired Memory: %.2f GB\n", $1 * $size / 1024 / 1024 / 1024;'
    else
      # Linux
      free -h
    fi
    
    # Show current configuration
    echo ""
    echo "Current Configuration"
    echo "--------------------"
    echo "GOMAXPROCS: ${GOMAXPROCS:-$(nproc)}"
    echo "GOGC: ${GOGC:-100}"
    echo "Memory Threshold: $MEMORY_THRESHOLD MB"
    echo "Memory Optimizer Enabled: $ENABLE_OPTIMIZER"
    echo "Resource Pool Manager Enabled: $ENABLE_POOL_MANAGER"
    echo "Configuration Tuner Enabled: $ENABLE_TUNER"
    ;;
    
  optimize)
    echo "Running Memory Optimization"
    echo "--------------------------"
    
    # Run memory optimization
    go run ./cmd/memory-benchmark/main.go \
      --templates=100 \
      --size=10000 \
      --variables=20 \
      --concurrent=10 \
      --iterations=1 \
      --optimizer=true \
      --poolmanager=false \
      --tuner=false \
      --output="$CONFIG_OUTPUT/optimization-report.md" \
      --verbose=true
    
    echo "Optimization complete. Report saved to $CONFIG_OUTPUT/optimization-report.md"
    ;;
    
  tune)
    echo "Running Configuration Tuning"
    echo "--------------------------"
    
    # Run configuration tuning
    go run ./cmd/memory-benchmark/main.go \
      --templates=100 \
      --size=10000 \
      --variables=20 \
      --concurrent=10 \
      --iterations=1 \
      --optimizer=false \
      --poolmanager=false \
      --tuner=true \
      --output="$CONFIG_OUTPUT/tuning-report.md" \
      --verbose=true
    
    echo "Tuning complete. Report saved to $CONFIG_OUTPUT/tuning-report.md"
    ;;
    
  profile)
    echo "Capturing Memory Profile"
    echo "----------------------"
    
    # Capture memory profile
    TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
    go run ./cmd/memory-benchmark/main.go \
      --templates=100 \
      --size=10000 \
      --variables=20 \
      --concurrent=10 \
      --iterations=1 \
      --optimizer=false \
      --poolmanager=false \
      --tuner=false \
      --output="$PROFILE_OUTPUT/profile-report-$TIMESTAMP.md" \
      --verbose=true
    
    echo "Profile captured. Report saved to $PROFILE_OUTPUT/profile-report-$TIMESTAMP.md"
    ;;
    
  benchmark)
    echo "Running Memory Benchmark"
    echo "----------------------"
    
    # Run memory benchmark
    TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
    go run ./cmd/memory-benchmark/main.go \
      --templates=200 \
      --size=10000 \
      --variables=30 \
      --concurrent=20 \
      --iterations=3 \
      --optimizer=$ENABLE_OPTIMIZER \
      --poolmanager=$ENABLE_POOL_MANAGER \
      --tuner=$ENABLE_TUNER \
      --output="$CONFIG_OUTPUT/benchmark-report-$TIMESTAMP.md" \
      --verbose=true
    
    echo "Benchmark complete. Report saved to $CONFIG_OUTPUT/benchmark-report-$TIMESTAMP.md"
    ;;
    
  *)
    echo "Unknown action: $ACTION"
    echo "Use --help for usage information"
    exit 1
    ;;
esac
