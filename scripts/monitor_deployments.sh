#!/bin/bash

# Deployment Monitoring Script for LLM Red Team v0.2.0
# This script monitors active deployments and collects metrics

set -e

# Configuration
TRACKER_HOST=${TRACKER_HOST:-"localhost:8091"}
CHECK_INTERVAL=${CHECK_INTERVAL:-300} # 5 minutes
ALERT_WEBHOOK=${ALERT_WEBHOOK:-""}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}=== LLM Red Team v0.2.0 Deployment Monitor ===${NC}"
echo "Monitoring deployments every ${CHECK_INTERVAL} seconds"
echo

# Function to check deployment health
check_deployment_health() {
    local deployment_id=$1
    local api_endpoint=$2
    
    echo -e "${BLUE}Checking deployment: ${deployment_id}${NC}"
    
    # Get metrics from deployment
    metrics=$(curl -s "${api_endpoint}/api/v1/metrics" 2>/dev/null || echo "{}")
    
    if [ -z "$metrics" ] || [ "$metrics" = "{}" ]; then
        echo -e "${RED}✗ Failed to retrieve metrics from ${deployment_id}${NC}"
        report_issue "$deployment_id" "P1" "connectivity" "Failed to retrieve metrics"
        return 1
    fi
    
    # Parse metrics using Python (more reliable than bash for JSON)
    python3 -c "
import json
import sys

try:
    data = json.loads('$metrics')
    perf = data.get('performance', {})
    
    success_rate = perf.get('success_rate', 0)
    response_time = perf.get('avg_response_time', 0)
    concurrent = perf.get('concurrent_attacks', 0)
    
    print(f'{success_rate},{response_time},{concurrent}')
except:
    print('0,0,0')
    sys.exit(1)
" | IFS=',' read -r success_rate response_time concurrent_attacks
    
    # Update metrics in tracker
    update_metrics "$deployment_id" "$success_rate" "$response_time" "$concurrent_attacks"
    
    # Check thresholds
    if (( $(echo "$success_rate < 0.95" | bc -l) )); then
        echo -e "${YELLOW}⚠ Low success rate: ${success_rate}${NC}"
        report_issue "$deployment_id" "P1" "performance" "Success rate below 95%: ${success_rate}"
    fi
    
    if (( $(echo "$response_time > 2.0" | bc -l) )); then
        echo -e "${YELLOW}⚠ High response time: ${response_time}s${NC}"
        report_issue "$deployment_id" "P2" "performance" "Response time above 2s: ${response_time}s"
    fi
    
    echo -e "${GREEN}✓ Health check completed${NC}"
}

# Function to update metrics in tracker
update_metrics() {
    local deployment_id=$1
    local success_rate=$2
    local response_time=$3
    local concurrent=$4
    
    # Get current metrics for calculating totals
    current_metrics=$(curl -s "${TRACKER_HOST}/api/v1/deployments/status" | \
        python3 -c "
import json, sys
data = json.load(sys.stdin)
deployment = data.get('$deployment_id', {})
metrics = deployment.get('metrics', {})
print(metrics.get('total_requests', 0))
")
    
    # Simulate request count increase (in production, get actual values)
    total_requests=$((current_metrics + 1000))
    error_count=$(echo "$total_requests * (1 - $success_rate)" | bc -l | cut -d. -f1)
    
    # Update metrics
    curl -s -X POST "${TRACKER_HOST}/api/v1/deployments/metrics?deployment_id=${deployment_id}" \
        -H "Content-Type: application/json" \
        -d "{
            \"total_requests\": ${total_requests},
            \"success_rate\": ${success_rate},
            \"avg_response_time\": ${response_time},
            \"peak_concurrent\": ${concurrent},
            \"error_count\": ${error_count}
        }" > /dev/null
}

# Function to report issues
report_issue() {
    local deployment_id=$1
    local severity=$2
    local issue_type=$3
    local description=$4
    
    curl -s -X POST "${TRACKER_HOST}/api/v1/deployments/issues?deployment_id=${deployment_id}" \
        -H "Content-Type: application/json" \
        -d "{
            \"severity\": \"${severity}\",
            \"type\": \"${issue_type}\",
            \"description\": \"${description}\"
        }" > /dev/null
    
    # Send alert for P0/P1 issues
    if [ "$severity" = "P0" ] || [ "$severity" = "P1" ]; then
        send_alert "$deployment_id" "$severity" "$description"
    fi
}

# Function to send alerts
send_alert() {
    local deployment_id=$1
    local severity=$2
    local description=$3
    
    if [ -n "$ALERT_WEBHOOK" ]; then
        curl -s -X POST "$ALERT_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{
                \"text\": \"🚨 ${severity} Alert for ${deployment_id}: ${description}\"
            }" > /dev/null
    fi
    
    echo -e "${RED}🚨 ALERT: ${severity} issue for ${deployment_id}: ${description}${NC}"
}

# Function to display summary
display_summary() {
    echo
    echo -e "${GREEN}=== Deployment Health Summary ===${NC}"
    
    health=$(curl -s "${TRACKER_HOST}/api/v1/deployments/health")
    
    echo "$health" | python3 -c "
import json, sys
data = json.load(sys.stdin)

print(f\"Total Deployments: {data.get('total_deployments', 0)}\")
print(f\"Healthy Deployments: {data.get('healthy_deployments', 0)}\")
print(f\"Health Percentage: {data.get('health_percentage', 0):.1f}%\")
print(f\"Total Requests: {data.get('total_requests', 0):,}\")
print(f\"Overall Success Rate: {data.get('overall_success_rate', 0):.2%}\")
print(f\"Open Issues: {data.get('open_issues', 0)}\")
"
    echo
}

# Function to register test deployments (for demo purposes)
register_test_deployments() {
    echo -e "${BLUE}Registering test deployments...${NC}"
    
    # Register test deployment
    curl -s -X POST "${TRACKER_HOST}/api/v1/deployments/register" \
        -H "Content-Type: application/json" \
        -d '{
            "organization": "Test Organization",
            "version": "v0.2.0",
            "environment": "test"
        }' > /dev/null
    
    echo -e "${GREEN}✓ Test deployment registered${NC}"
}

# Main monitoring loop
main() {
    # Check if tracker is running
    if ! curl -s "${TRACKER_HOST}/api/v1/deployments/health" > /dev/null 2>&1; then
        echo -e "${RED}Error: Deployment tracker not running at ${TRACKER_HOST}${NC}"
        echo "Start the tracker with: go run deployment-tracker/tracker.go"
        exit 1
    fi
    
    # For demo, register a test deployment if none exist
    deployments=$(curl -s "${TRACKER_HOST}/api/v1/deployments/status")
    if [ "$deployments" = "{}" ]; then
        register_test_deployments
    fi
    
    while true; do
        echo -e "${BLUE}$(date '+%Y-%m-%d %H:%M:%S') - Starting health checks${NC}"
        
        # Get all deployments
        deployments=$(curl -s "${TRACKER_HOST}/api/v1/deployments/status")
        
        # Check each deployment
        echo "$deployments" | python3 -c "
import json, sys
data = json.load(sys.stdin)
for dep_id, info in data.items():
    print(f\"{dep_id},{info.get('organization', 'Unknown')}\")" | \
        while IFS=',' read -r deployment_id organization; do
            # In production, this would be the actual deployment endpoint
            # For now, we'll use the local API endpoint as a simulation
            api_endpoint="http://localhost:8080"
            
            check_deployment_health "$deployment_id" "$api_endpoint"
        done
        
        # Display summary
        display_summary
        
        # Wait for next check
        echo -e "${BLUE}Next check in ${CHECK_INTERVAL} seconds...${NC}"
        sleep "$CHECK_INTERVAL"
    done
}

# Handle cleanup on exit
cleanup() {
    echo
    echo -e "${YELLOW}Monitoring stopped${NC}"
}

trap cleanup EXIT

# Parse arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --interval SECONDS  Check interval (default: 300)"
        echo "  --tracker HOST:PORT Tracker endpoint (default: localhost:8091)"
        echo "  --webhook URL       Alert webhook URL"
        echo
        echo "Environment variables:"
        echo "  TRACKER_HOST        Tracker endpoint"
        echo "  CHECK_INTERVAL      Check interval in seconds"
        echo "  ALERT_WEBHOOK       Webhook URL for alerts"
        exit 0
        ;;
    --interval)
        CHECK_INTERVAL="$2"
        shift 2
        ;;
    --tracker)
        TRACKER_HOST="$2"
        shift 2
        ;;
    --webhook)
        ALERT_WEBHOOK="$2"
        shift 2
        ;;
    *)
        main
        ;;
esac