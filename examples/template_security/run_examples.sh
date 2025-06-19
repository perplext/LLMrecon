#!/bin/bash

# Script to demonstrate the template security framework

# Create necessary directories
mkdir -p logs
mkdir -p template_storage
mkdir -p dashboard_templates

# Build the example
echo "Building template security example..."
go build -o template-security main.go dashboard.go

# Test safe template
echo -e "\n\n===== Testing Safe Template ====="
./template-security --template=sample_templates/safe_template.tmpl --verbose

# Test risky template with strict mode
echo -e "\n\n===== Testing Risky Template (Strict Mode) ====="
./template-security --template=sample_templates/risky_template.tmpl --mode=strict --verbose

# Test risky template with audit mode
echo -e "\n\n===== Testing Risky Template (Audit Mode) ====="
./template-security --template=sample_templates/risky_template.tmpl --mode=audit --verbose

# Test with workflow
echo -e "\n\n===== Testing Template Approval Workflow ====="
./template-security --template=sample_templates/safe_template.tmpl --workflow --user=admin --verbose

# Test validation only
echo -e "\n\n===== Testing Validation Only ====="
./template-security --template=sample_templates/risky_template.tmpl --validate --execute=false --verbose

# Test batch processing
echo -e "\n\n===== Testing Batch Processing ====="
./template-security --batch --template-dir=sample_templates --verbose

# Start dashboard in background
echo -e "\n\n===== Starting Dashboard ====="
./template-security --dashboard --port=8080 &
DASHBOARD_PID=$!

# Wait a moment for the dashboard to start
sleep 2

# Open the dashboard in the default browser
echo "Opening dashboard in browser..."
open http://localhost:8080

# Run some templates to generate metrics for the dashboard
echo "Generating metrics for the dashboard..."
./template-security --template=sample_templates/safe_template.tmpl --verbose
./template-security --template=sample_templates/risky_template.tmpl --mode=audit --verbose

# Wait for user to view the dashboard
echo "Dashboard is running at http://localhost:8080"
echo "Press Enter to stop the dashboard and exit..."
read

# Kill the dashboard process
kill $DASHBOARD_PID

echo -e "\n\nAll examples completed. Check the logs directory for detailed logs."
echo "You can start the dashboard again with: ./template-security --dashboard --port=8080"
