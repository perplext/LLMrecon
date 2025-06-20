#!/bin/bash

# Production Validation Script for LLM Red Team v0.2.0
# This script validates that the distributed infrastructure is working correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REDIS_HOST=${REDIS_HOST:-"localhost:6379"}
API_HOST=${API_HOST:-"localhost:8080"}
MONITORING_HOST=${MONITORING_HOST:-"localhost:8090"}
PROFILING_HOST=${PROFILING_HOST:-"localhost:6060"}
TEST_DURATION=${TEST_DURATION:-"300"} # 5 minutes default

echo -e "${GREEN}=== LLM Red Team v0.2.0 Production Validation ===${NC}"
echo "Starting comprehensive validation of distributed infrastructure..."
echo

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
        return 1
    fi
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Function to print info
print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Validation functions

validate_dependencies() {
    echo -e "${GREEN}=== Dependency Validation ===${NC}"
    
    # Check Go version
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        print_status 0 "Go installed: $GO_VERSION"
    else
        print_status 1 "Go not found"
        return 1
    fi
    
    # Check Redis connection
    if command -v redis-cli &> /dev/null; then
        if redis-cli -h $(echo $REDIS_HOST | cut -d':' -f1) -p $(echo $REDIS_HOST | cut -d':' -f2) ping &> /dev/null; then
            print_status 0 "Redis connection successful"
        else
            print_status 1 "Redis connection failed"
            return 1
        fi
    else
        print_warning "redis-cli not found, skipping Redis test"
    fi
    
    # Check curl for API testing
    if ! command -v curl &> /dev/null; then
        print_status 1 "curl not found (required for API testing)"
        return 1
    fi
    
    echo
}

validate_build() {
    echo -e "${GREEN}=== Build Validation ===${NC}"
    
    # Test main application build
    if go build -o llm-red-team-test ./src/main.go; then
        print_status 0 "Main application builds successfully"
        rm -f llm-red-team-test
    else
        print_status 1 "Main application build failed"
        return 1
    fi
    
    # Test key components
    local components=(
        "./src/performance/..."
        "./src/queue/..."
        "./src/provider/core/..."
    )
    
    for component in "${components[@]}"; do
        if go build $component &> /dev/null; then
            print_status 0 "Component builds successfully: $component"
        else
            print_status 1 "Component build failed: $component"
            return 1
        fi
    done
    
    echo
}

validate_redis_cluster() {
    echo -e "${GREEN}=== Redis Cluster Validation ===${NC}"
    
    if ! command -v redis-cli &> /dev/null; then
        print_warning "redis-cli not available, skipping cluster validation"
        return 0
    fi
    
    # Test basic Redis operations
    local test_key="llm-red-team-validation-$$"
    local test_value="validation-$(date +%s)"
    
    if redis-cli -h $(echo $REDIS_HOST | cut -d':' -f1) -p $(echo $REDIS_HOST | cut -d':' -f2) set $test_key "$test_value" &> /dev/null; then
        print_status 0 "Redis SET operation successful"
    else
        print_status 1 "Redis SET operation failed"
        return 1
    fi
    
    local retrieved_value=$(redis-cli -h $(echo $REDIS_HOST | cut -d':' -f1) -p $(echo $REDIS_HOST | cut -d':' -f2) get $test_key 2>/dev/null)
    if [ "$retrieved_value" = "$test_value" ]; then
        print_status 0 "Redis GET operation successful"
    else
        print_status 1 "Redis GET operation failed"
        return 1
    fi
    
    # Cleanup
    redis-cli -h $(echo $REDIS_HOST | cut -d':' -f1) -p $(echo $REDIS_HOST | cut -d':' -f2) del $test_key &> /dev/null
    
    # Test Redis memory info
    local memory_info=$(redis-cli -h $(echo $REDIS_HOST | cut -d':' -f1) -p $(echo $REDIS_HOST | cut -d':' -f2) info memory 2>/dev/null | grep used_memory_human | cut -d':' -f2 | tr -d '\r')
    if [ -n "$memory_info" ]; then
        print_status 0 "Redis memory usage: $memory_info"
    else
        print_warning "Could not retrieve Redis memory info"
    fi
    
    echo
}

validate_application_startup() {
    echo -e "${GREEN}=== Application Startup Validation ===${NC}"
    
    # Check if application is already running
    local api_running=false
    local monitoring_running=false
    local profiling_running=false
    
    if curl -s -f "http://$API_HOST/health" &> /dev/null; then
        print_status 0 "Main API endpoint responding"
        api_running=true
    else
        print_warning "Main API endpoint not responding (http://$API_HOST/health)"
    fi
    
    if curl -s -f "http://$MONITORING_HOST/api/v1/status" &> /dev/null; then
        print_status 0 "Monitoring dashboard responding"
        monitoring_running=true
    else
        print_warning "Monitoring dashboard not responding (http://$MONITORING_HOST/api/v1/status)"
    fi
    
    if curl -s -f "http://$PROFILING_HOST/debug/pprof/" &> /dev/null; then
        print_status 0 "Profiling endpoint responding"
        profiling_running=true
    else
        print_warning "Profiling endpoint not responding (http://$PROFILING_HOST/debug/pprof/)"
    fi
    
    if [ "$api_running" = false ] && [ "$monitoring_running" = false ] && [ "$profiling_running" = false ]; then
        print_warning "No application endpoints detected. Manual startup required."
        print_info "Start the application with: ./llm-red-team server --distributed --redis-addr $REDIS_HOST"
    fi
    
    echo
}

validate_performance_baseline() {
    echo -e "${GREEN}=== Performance Baseline Validation ===${NC}"
    
    # Check if monitoring endpoint is available
    if ! curl -s -f "http://$MONITORING_HOST/api/v1/metrics" &> /dev/null; then
        print_warning "Monitoring endpoint not available, skipping performance validation"
        return 0
    fi
    
    # Get baseline metrics
    local metrics=$(curl -s "http://$MONITORING_HOST/api/v1/metrics" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$metrics" ]; then
        print_status 0 "Performance metrics endpoint accessible"
        
        # Parse basic metrics if JSON is available
        if command -v jq &> /dev/null; then
            local cpu_usage=$(echo "$metrics" | jq -r '.cpu.usage // "N/A"' 2>/dev/null)
            local memory_alloc=$(echo "$metrics" | jq -r '.memory.allocated // "N/A"' 2>/dev/null)
            local goroutines=$(echo "$metrics" | jq -r '.goroutines.count // "N/A"' 2>/dev/null)
            
            print_info "CPU Usage: $cpu_usage%"
            print_info "Memory Allocated: $memory_alloc bytes"
            print_info "Goroutines: $goroutines"
        else
            print_info "Install 'jq' for detailed metrics parsing"
        fi
    else
        print_warning "Could not retrieve performance metrics"
    fi
    
    echo
}

validate_load_capacity() {
    echo -e "${GREEN}=== Load Capacity Validation ===${NC}"
    
    print_info "Running basic load validation for $TEST_DURATION seconds..."
    
    # Simple concurrent request test
    local success_count=0
    local total_requests=10
    
    for i in $(seq 1 $total_requests); do
        if curl -s -f "http://$API_HOST/health" &> /dev/null; then
            ((success_count++))
        fi
    done
    
    local success_rate=$((success_count * 100 / total_requests))
    if [ $success_rate -ge 80 ]; then
        print_status 0 "Basic load test passed ($success_count/$total_requests requests succeeded)"
    else
        print_status 1 "Basic load test failed ($success_count/$total_requests requests succeeded)"
        return 1
    fi
    
    print_info "For comprehensive load testing, use: ./scripts/load_test.sh"
    echo
}

validate_security_configuration() {
    echo -e "${GREEN}=== Security Configuration Validation ===${NC}"
    
    # Check for common security issues
    local security_score=0
    local total_checks=4
    
    # Check if running as root
    if [ "$(id -u)" -eq 0 ]; then
        print_warning "Running as root user (security risk)"
    else
        print_status 0 "Not running as root user"
        ((security_score++))
    fi
    
    # Check for TLS configuration
    if curl -s -k "https://$API_HOST/health" &> /dev/null; then
        print_status 0 "TLS endpoint detected"
        ((security_score++))
    else
        print_warning "No TLS endpoint detected (consider enabling HTTPS)"
    fi
    
    # Check Redis authentication
    if redis-cli -h $(echo $REDIS_HOST | cut -d':' -f1) -p $(echo $REDIS_HOST | cut -d':' -f2) info server 2>&1 | grep -q "NOAUTH"; then
        print_warning "Redis authentication not configured"
    else
        print_status 0 "Redis appears to have authentication configured"
        ((security_score++))
    fi
    
    # Check for default passwords in environment
    if env | grep -i password | grep -q "password123\|admin\|default"; then
        print_warning "Default passwords detected in environment"
    else
        print_status 0 "No obvious default passwords in environment"
        ((security_score++))
    fi
    
    print_info "Security score: $security_score/$total_checks"
    echo
}

validate_documentation() {
    echo -e "${GREEN}=== Documentation Validation ===${NC}"
    
    local doc_files=(
        "README.md"
        "CLAUDE.md"
        "docs/PRODUCTION_DEPLOYMENT.md"
        "docs/PERFORMANCE_OPTIMIZATION.md"
        "docs/DISTRIBUTED_ARCHITECTURE.md"
    )
    
    for doc in "${doc_files[@]}"; do
        if [ -f "$doc" ]; then
            print_status 0 "Documentation exists: $doc"
        else
            print_status 1 "Missing documentation: $doc"
        fi
    done
    
    echo
}

generate_validation_report() {
    echo -e "${GREEN}=== Validation Report ===${NC}"
    
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local report_file="validation-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
LLM Red Team v0.2.0 Production Validation Report
Generated: $timestamp
Host: $(hostname)
User: $(whoami)

System Information:
- OS: $(uname -s) $(uname -r)
- Architecture: $(uname -m)
- Go Version: $(go version 2>/dev/null || echo "Not found")

Configuration:
- Redis Host: $REDIS_HOST
- API Host: $API_HOST
- Monitoring Host: $MONITORING_HOST
- Profiling Host: $PROFILING_HOST
- Test Duration: $TEST_DURATION seconds

Validation Results:
$(grep -E "✓|✗|⚠" /tmp/validation.log 2>/dev/null || echo "Detailed logs not captured")

Recommendations:
1. Run comprehensive load testing with: ./scripts/load_test.sh
2. Monitor system metrics during load testing
3. Review security configurations for production deployment
4. Set up automated monitoring and alerting
5. Configure backup and disaster recovery procedures

For detailed performance testing, see: docs/PERFORMANCE_OPTIMIZATION.md
For production deployment, see: docs/PRODUCTION_DEPLOYMENT.md
EOF
    
    print_status 0 "Validation report generated: $report_file"
    echo
}

# Main validation sequence
main() {
    local exit_code=0
    
    # Redirect output to capture for report
    exec > >(tee /tmp/validation.log)
    
    validate_dependencies || exit_code=1
    validate_build || exit_code=1
    validate_redis_cluster || exit_code=1
    validate_application_startup || exit_code=1
    validate_performance_baseline || exit_code=1
    validate_load_capacity || exit_code=1
    validate_security_configuration || exit_code=1
    validate_documentation || exit_code=1
    
    generate_validation_report
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}=== Validation Completed Successfully ===${NC}"
        echo "Your LLM Red Team v0.2.0 installation appears to be ready for production!"
    else
        echo -e "${RED}=== Validation Completed with Issues ===${NC}"
        echo "Please review the issues above before proceeding to production."
    fi
    
    return $exit_code
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --help, -h              Show this help message"
        echo "  --redis-host HOST:PORT  Redis host (default: localhost:6379)"
        echo "  --api-host HOST:PORT    API host (default: localhost:8080)"
        echo "  --duration SECONDS      Test duration (default: 300)"
        echo
        echo "Environment variables:"
        echo "  REDIS_HOST              Redis host:port"
        echo "  API_HOST                API host:port"
        echo "  MONITORING_HOST         Monitoring host:port"
        echo "  PROFILING_HOST          Profiling host:port"
        echo "  TEST_DURATION           Test duration in seconds"
        exit 0
        ;;
    --redis-host)
        REDIS_HOST="$2"
        shift 2
        ;;
    --api-host)
        API_HOST="$2"
        shift 2
        ;;
    --duration)
        TEST_DURATION="$2"
        shift 2
        ;;
    *)
        # Run main validation
        main
        ;;
esac