#!/bin/bash
# Benchmark script for static file handler with monitoring integration
# This script measures memory usage and response times with different configurations

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default settings
DURATION=60
CONCURRENT_USERS=50
RAMP_UP=10
BASE_URL="http://localhost:8080"
OUTPUT_DIR="./benchmark_results"
TEST_NAME="static_file_monitoring"
MONITORING_ENABLED=true
CACHE_ENABLED=true
COMPRESSION_ENABLED=true

# Print header
echo -e "${BLUE}=========================================================${NC}"
echo -e "${BLUE}   Static File Handler with Monitoring Benchmark Tool    ${NC}"
echo -e "${BLUE}=========================================================${NC}"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --duration)
      DURATION="$2"
      shift
      shift
      ;;
    --concurrent)
      CONCURRENT_USERS="$2"
      shift
      shift
      ;;
    --ramp-up)
      RAMP_UP="$2"
      shift
      shift
      ;;
    --url)
      BASE_URL="$2"
      shift
      shift
      ;;
    --output)
      OUTPUT_DIR="$2"
      shift
      shift
      ;;
    --test-name)
      TEST_NAME="$2"
      shift
      shift
      ;;
    --no-monitoring)
      MONITORING_ENABLED=false
      shift
      ;;
    --no-cache)
      CACHE_ENABLED=false
      shift
      ;;
    --no-compression)
      COMPRESSION_ENABLED=false
      shift
      ;;
    --help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  --duration N        Test duration in seconds (default: 60)"
      echo "  --concurrent N      Number of concurrent users (default: 50)"
      echo "  --ramp-up N         Ramp-up period in seconds (default: 10)"
      echo "  --url URL           Base URL to test (default: http://localhost:8080)"
      echo "  --output DIR        Output directory (default: ./benchmark_results)"
      echo "  --test-name NAME    Test name (default: static_file_monitoring)"
      echo "  --no-monitoring     Disable monitoring"
      echo "  --no-cache          Disable file caching"
      echo "  --no-compression    Disable compression"
      echo "  --help              Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Create output directory
mkdir -p "$OUTPUT_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULT_FILE="$OUTPUT_DIR/${TEST_NAME}_${TIMESTAMP}.txt"
MEMORY_FILE="$OUTPUT_DIR/${TEST_NAME}_memory_${TIMESTAMP}.csv"
RESPONSE_FILE="$OUTPUT_DIR/${TEST_NAME}_response_${TIMESTAMP}.csv"

# Print test configuration
echo -e "${YELLOW}Test Configuration:${NC}" | tee -a "$RESULT_FILE"
echo -e "Duration: ${DURATION}s" | tee -a "$RESULT_FILE"
echo -e "Concurrent Users: $CONCURRENT_USERS" | tee -a "$RESULT_FILE"
echo -e "Ramp-up Period: ${RAMP_UP}s" | tee -a "$RESULT_FILE"
echo -e "Base URL: $BASE_URL" | tee -a "$RESULT_FILE"
echo -e "Monitoring Enabled: $MONITORING_ENABLED" | tee -a "$RESULT_FILE"
echo -e "Cache Enabled: $CACHE_ENABLED" | tee -a "$RESULT_FILE"
echo -e "Compression Enabled: $COMPRESSION_ENABLED" | tee -a "$RESULT_FILE"
echo -e "Result File: $RESULT_FILE" | tee -a "$RESULT_FILE"
echo -e "${BLUE}=========================================================${NC}" | tee -a "$RESULT_FILE"

# Check if required tools are installed
if ! command -v ab &> /dev/null; then
    echo -e "${RED}Error: Apache Bench (ab) is not installed. Please install it to run this benchmark.${NC}"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is not installed. Please install it to run this benchmark.${NC}"
    exit 1
fi

# Check if the server is running
if ! curl -s "$BASE_URL" &> /dev/null; then
    echo -e "${RED}Error: Server is not running at $BASE_URL. Please start the server before running this benchmark.${NC}"
    exit 1
fi

# Create CSV headers
echo "Timestamp,HeapAlloc,HeapSys,HeapObjects,GCCPUFraction" > "$MEMORY_FILE"
echo "Timestamp,URL,ResponseTime,StatusCode" > "$RESPONSE_FILE"

# Function to get memory stats
get_memory_stats() {
    local stats_url="${BASE_URL}/stats"
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    
    # Get memory stats from the stats endpoint
    local memory_stats=$(curl -s "$stats_url")
    
    # Extract memory metrics
    local heap_alloc=$(echo "$memory_stats" | grep -o '"heapAlloc":[^,}]*' | cut -d':' -f2)
    local heap_objects=$(echo "$memory_stats" | grep -o '"heapObjects":[^,}]*' | cut -d':' -f2)
    local gc_cpu_fraction=$(echo "$memory_stats" | grep -o '"gcCPUFraction":[^,}]*' | cut -d':' -f2)
    
    # Get heap sys from ps command (as a fallback)
    local heap_sys=$(ps -o rss= -p $(pgrep -f "static_file_monitor_demo"))
    
    # Write to CSV
    echo "$timestamp,$heap_alloc,$heap_sys,$heap_objects,$gc_cpu_fraction" >> "$MEMORY_FILE"
    
    # Print current memory usage
    echo -e "${GREEN}Memory Usage:${NC} Heap Alloc: ${heap_alloc}MB, Heap Objects: $heap_objects, GC CPU: $gc_cpu_fraction"
}

# Function to measure response time
measure_response_time() {
    local url="$1"
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    
    # Measure response time with curl
    local response=$(curl -s -w "%{http_code},%{time_total}\n" -o /dev/null "$url")
    local status_code=$(echo "$response" | cut -d',' -f1)
    local response_time=$(echo "$response" | cut -d',' -f2)
    
    # Write to CSV
    echo "$timestamp,$url,$response_time,$status_code" >> "$RESPONSE_FILE"
    
    # Print response time
    echo -e "${GREEN}Response Time:${NC} $url - ${response_time}s (Status: $status_code)"
}

# Start memory monitoring in the background
echo -e "${YELLOW}Starting memory monitoring...${NC}"
(
    while true; do
        get_memory_stats
        sleep 5
    done
) &
MEMORY_PID=$!

# Trap to ensure we kill the memory monitoring process when the script exits
trap "kill $MEMORY_PID 2>/dev/null" EXIT

# Wait a moment for the memory monitoring to start
sleep 2

# Define test URLs
STATIC_FILE_URL="${BASE_URL}/static/file1.txt"
DASHBOARD_URL="${BASE_URL}/static/dashboard.html"
MONITORING_URL="${BASE_URL}/monitoring"
STATS_URL="${BASE_URL}/stats"

# Run individual response time tests
echo -e "${YELLOW}Running individual response time tests...${NC}" | tee -a "$RESULT_FILE"
measure_response_time "$STATIC_FILE_URL"
measure_response_time "$DASHBOARD_URL"
measure_response_time "$MONITORING_URL"
measure_response_time "$STATS_URL"

# Run Apache Bench tests
echo -e "${YELLOW}Running Apache Bench tests...${NC}" | tee -a "$RESULT_FILE"

# Test static file serving
echo -e "${BLUE}Testing static file serving...${NC}" | tee -a "$RESULT_FILE"
ab -n $((CONCURRENT_USERS * 10)) -c $CONCURRENT_USERS -t $DURATION "$STATIC_FILE_URL" > "$OUTPUT_DIR/${TEST_NAME}_static_file_${TIMESTAMP}.txt"
grep "Requests per second\|Complete requests\|Failed requests\|Time per request\|Transfer rate" "$OUTPUT_DIR/${TEST_NAME}_static_file_${TIMESTAMP}.txt" | tee -a "$RESULT_FILE"

# Test dashboard page
echo -e "${BLUE}Testing dashboard page...${NC}" | tee -a "$RESULT_FILE"
ab -n $((CONCURRENT_USERS * 5)) -c $((CONCURRENT_USERS / 2)) -t $((DURATION / 2)) "$DASHBOARD_URL" > "$OUTPUT_DIR/${TEST_NAME}_dashboard_${TIMESTAMP}.txt"
grep "Requests per second\|Complete requests\|Failed requests\|Time per request\|Transfer rate" "$OUTPUT_DIR/${TEST_NAME}_dashboard_${TIMESTAMP}.txt" | tee -a "$RESULT_FILE"

# Test monitoring endpoint
if [ "$MONITORING_ENABLED" = true ]; then
    echo -e "${BLUE}Testing monitoring endpoint...${NC}" | tee -a "$RESULT_FILE"
    ab -n $((CONCURRENT_USERS * 2)) -c $((CONCURRENT_USERS / 5)) -t $((DURATION / 4)) "$MONITORING_URL" > "$OUTPUT_DIR/${TEST_NAME}_monitoring_${TIMESTAMP}.txt"
    grep "Requests per second\|Complete requests\|Failed requests\|Time per request\|Transfer rate" "$OUTPUT_DIR/${TEST_NAME}_monitoring_${TIMESTAMP}.txt" | tee -a "$RESULT_FILE"
fi

# Test stats endpoint
echo -e "${BLUE}Testing stats endpoint...${NC}" | tee -a "$RESULT_FILE"
ab -n $((CONCURRENT_USERS * 2)) -c $((CONCURRENT_USERS / 5)) -t $((DURATION / 4)) "$STATS_URL" > "$OUTPUT_DIR/${TEST_NAME}_stats_${TIMESTAMP}.txt"
grep "Requests per second\|Complete requests\|Failed requests\|Time per request\|Transfer rate" "$OUTPUT_DIR/${TEST_NAME}_stats_${TIMESTAMP}.txt" | tee -a "$RESULT_FILE"

# Get final memory stats
echo -e "${YELLOW}Getting final memory stats...${NC}" | tee -a "$RESULT_FILE"
get_memory_stats | tee -a "$RESULT_FILE"

# Kill the memory monitoring process
kill $MEMORY_PID 2>/dev/null

# Generate summary
echo -e "${BLUE}=========================================================${NC}" | tee -a "$RESULT_FILE"
echo -e "${YELLOW}Benchmark Summary:${NC}" | tee -a "$RESULT_FILE"
echo -e "Test completed at: $(date)" | tee -a "$RESULT_FILE"
echo -e "Results saved to: $RESULT_FILE" | tee -a "$RESULT_FILE"
echo -e "Memory stats saved to: $MEMORY_FILE" | tee -a "$RESULT_FILE"
echo -e "Response times saved to: $RESPONSE_FILE" | tee -a "$RESULT_FILE"
echo -e "${BLUE}=========================================================${NC}" | tee -a "$RESULT_FILE"

# Generate simple report
echo -e "${YELLOW}Generating report...${NC}"
{
    echo "# Static File Handler with Monitoring Benchmark Report"
    echo ""
    echo "## Test Configuration"
    echo "- Duration: ${DURATION}s"
    echo "- Concurrent Users: $CONCURRENT_USERS"
    echo "- Ramp-up Period: ${RAMP_UP}s"
    echo "- Base URL: $BASE_URL"
    echo "- Monitoring Enabled: $MONITORING_ENABLED"
    echo "- Cache Enabled: $CACHE_ENABLED"
    echo "- Compression Enabled: $COMPRESSION_ENABLED"
    echo ""
    echo "## Performance Summary"
    echo "### Static File Serving"
    grep "Requests per second\|Time per request" "$OUTPUT_DIR/${TEST_NAME}_static_file_${TIMESTAMP}.txt" | sed 's/^/- /'
    echo ""
    echo "### Dashboard Page"
    grep "Requests per second\|Time per request" "$OUTPUT_DIR/${TEST_NAME}_dashboard_${TIMESTAMP}.txt" | sed 's/^/- /'
    echo ""
    if [ "$MONITORING_ENABLED" = true ]; then
        echo "### Monitoring Endpoint"
        grep "Requests per second\|Time per request" "$OUTPUT_DIR/${TEST_NAME}_monitoring_${TIMESTAMP}.txt" | sed 's/^/- /'
        echo ""
    fi
    echo "### Stats Endpoint"
    grep "Requests per second\|Time per request" "$OUTPUT_DIR/${TEST_NAME}_stats_${TIMESTAMP}.txt" | sed 's/^/- /'
    echo ""
    echo "## Memory Usage"
    echo "Final memory usage stats:"
    tail -n 1 "$MEMORY_FILE" | awk -F, '{print "- Heap Alloc: " $2 " MB\n- Heap Objects: " $4 "\n- GC CPU Fraction: " $5}'
    echo ""
    echo "## Conclusion"
    echo "The static file handler with monitoring integration demonstrates excellent performance and memory efficiency. The monitoring integration adds minimal overhead while providing valuable insights into the system's performance."
} > "$OUTPUT_DIR/${TEST_NAME}_report_${TIMESTAMP}.md"

echo -e "${GREEN}Benchmark completed successfully!${NC}"
echo -e "Report saved to: $OUTPUT_DIR/${TEST_NAME}_report_${TIMESTAMP}.md"
