#!/bin/bash

# Script to build specific components of the codebase
# Usage: ./build_component.sh <component_name>
# Example: ./build_component.sh template_security

set -e

COMPONENT=$1
BUILD_DIR="./build"
SRC_DIR="./src"
TEMP_DIR="/tmp/LLMrecon-build"

if [ -z "$COMPONENT" ]; then
    echo "Usage: ./build_component.sh <component_name>"
    echo "Available components:"
    echo "  template_security - Template security verification tool"
    echo "  audit_logger - Audit logging functionality"
    echo "  memory_optimizer - Memory optimization components"
    echo "  monitoring - Monitoring and metrics components"
    exit 1
fi

# Create build directories
mkdir -p "$BUILD_DIR"
mkdir -p "$TEMP_DIR"

echo "Building component: $COMPONENT"

case "$COMPONENT" in
    template_security)
        echo "Building template security component..."
        
        # Create a temporary directory with only the necessary files
        rm -rf "$TEMP_DIR"
        mkdir -p "$TEMP_DIR/src/cmd"
        mkdir -p "$TEMP_DIR/src/template/security"
        mkdir -p "$TEMP_DIR/src/testing/owasp/compliance"
        
        # Copy the necessary files
        cp -r "$SRC_DIR/cmd/root.go" "$TEMP_DIR/src/cmd/"
        cp -r "$SRC_DIR/cmd/template_security_simplified.go" "$TEMP_DIR/src/cmd/template_security.go"
        cp -r "$SRC_DIR/template/security" "$TEMP_DIR/src/template/"
        cp -r "$SRC_DIR/testing/owasp/compliance" "$TEMP_DIR/src/testing/owasp/"
        cp -r "$SRC_DIR/testing/owasp/types" "$TEMP_DIR/src/testing/owasp/"
        
        # Build the component
        cd "$TEMP_DIR"
        go build -o "$OLDPWD/$BUILD_DIR/template_security" ./src/cmd/root.go ./src/cmd/template_security.go
        
        echo "Template security component built successfully: $BUILD_DIR/template_security"
        ;;
        
    audit_logger)
        echo "Building audit logger component..."
        
        # Create a temporary directory with only the necessary files
        rm -rf "$TEMP_DIR"
        mkdir -p "$TEMP_DIR/src/audit"
        mkdir -p "$TEMP_DIR/src/security/audit"
        
        # Copy the necessary files
        cp -r "$SRC_DIR/audit" "$TEMP_DIR/src/"
        cp -r "$SRC_DIR/security/audit" "$TEMP_DIR/src/security/"
        
        # Build a simple test program to verify the audit logger
        cat > "$TEMP_DIR/src/audit_test.go" << EOF
package main

import (
	"context"
	"fmt"
	"os"
	
	"github.com/LLMrecon/LLMrecon/src/audit"
	secaudit "github.com/LLMrecon/LLMrecon/src/security/audit"
)

func main() {
	ctx := context.Background()
	
	// Initialize audit logger
	logger, err := audit.NewAuditLogger("test", "./logs")
	if err != nil {
		fmt.Printf("Error creating audit logger: %v\n", err)
		os.Exit(1)
	}
	
	// Log an event
	err = logger.LogEvent(ctx, &audit.Event{
		Type:        "test_event",
		Source:      "audit_test",
		Description: "Test audit event",
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	})
	
	if err != nil {
		fmt.Printf("Error logging event: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize security audit logger
	secLogger, err := secaudit.NewSecurityAuditLogger("security_test", "./logs")
	if err != nil {
		fmt.Printf("Error creating security audit logger: %v\n", err)
		os.Exit(1)
	}
	
	// Log a security event
	err = secLogger.LogSecurityEvent(ctx, &secaudit.SecurityEvent{
		Type:        "security_test_event",
		Source:      "audit_test",
		Description: "Test security audit event",
		Severity:    "medium",
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	})
	
	if err != nil {
		fmt.Printf("Error logging security event: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Audit logger test completed successfully")
}
EOF
        
        # Build the component
        cd "$TEMP_DIR"
        go build -o "$OLDPWD/$BUILD_DIR/audit_logger_test" ./src/audit_test.go
        
        echo "Audit logger component built successfully: $BUILD_DIR/audit_logger_test"
        ;;
        
    memory_optimizer)
        echo "Building memory optimizer component..."
        
        # Create a temporary directory with only the necessary files
        rm -rf "$TEMP_DIR"
        mkdir -p "$TEMP_DIR/src/utils/config"
        mkdir -p "$TEMP_DIR/src/utils/profiling"
        mkdir -p "$TEMP_DIR/src/utils/resource"
        mkdir -p "$TEMP_DIR/src/template/management/optimization"
        
        # Copy the necessary files
        cp -r "$SRC_DIR/utils/config" "$TEMP_DIR/src/utils/"
        cp -r "$SRC_DIR/utils/profiling" "$TEMP_DIR/src/utils/"
        cp -r "$SRC_DIR/utils/resource" "$TEMP_DIR/src/utils/"
        cp -r "$SRC_DIR/template/management/optimization" "$TEMP_DIR/src/template/management/"
        
        # Build a simple test program
        cat > "$TEMP_DIR/src/memory_optimizer_test.go" << EOF
package main

import (
	"fmt"
	"os"
	
	"github.com/LLMrecon/LLMrecon/src/utils/config"
	"github.com/LLMrecon/LLMrecon/src/utils/profiling"
	"github.com/LLMrecon/LLMrecon/src/utils/resource"
	"github.com/LLMrecon/LLMrecon/src/template/management/optimization"
)

func main() {
	// Load memory configuration
	cfg, err := config.LoadMemoryConfig("dev")
	if err != nil {
		fmt.Printf("Error loading memory config: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize memory profiler
	profiler, err := profiling.NewMemoryProfiler(cfg)
	if err != nil {
		fmt.Printf("Error creating memory profiler: %v\n", err)
		os.Exit(1)
	}
	
	// Start profiling
	profiler.StartProfiling()
	
	// Initialize resource pool manager
	poolManager, err := resource.NewPoolManager(cfg)
	if err != nil {
		fmt.Printf("Error creating resource pool manager: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize memory optimizer
	optimizer, err := optimization.NewMemoryOptimizer(cfg, profiler, poolManager)
	if err != nil {
		fmt.Printf("Error creating memory optimizer: %v\n", err)
		os.Exit(1)
	}
	
	// Run optimization
	result, err := optimizer.Optimize()
	if err != nil {
		fmt.Printf("Error optimizing memory: %v\n", err)
		os.Exit(1)
	}
	
	// Print optimization results
	fmt.Printf("Memory optimization completed successfully\n")
	fmt.Printf("Memory reduction: %.2f%%\n", result.MemoryReductionPercentage)
	fmt.Printf("Optimized templates: %d\n", result.OptimizedTemplateCount)
	
	// Stop profiling
	profiler.StopProfiling()
}
EOF
        
        # Build the component
        cd "$TEMP_DIR"
        go build -o "$OLDPWD/$BUILD_DIR/memory_optimizer_test" ./src/memory_optimizer_test.go
        
        echo "Memory optimizer component built successfully: $BUILD_DIR/memory_optimizer_test"
        ;;
        
    monitoring)
        echo "Building monitoring component..."
        
        # Create a temporary directory with only the necessary files
        rm -rf "$TEMP_DIR"
        mkdir -p "$TEMP_DIR/src/utils/monitoring"
        
        # Copy the necessary files
        cp -r "$SRC_DIR/utils/monitoring" "$TEMP_DIR/src/utils/"
        
        # Build a simple test program
        cat > "$TEMP_DIR/src/monitoring_test.go" << EOF
package main

import (
	"fmt"
	"os"
	"time"
	
	"github.com/LLMrecon/LLMrecon/src/utils/monitoring"
)

func main() {
	// Initialize monitoring service
	service, err := monitoring.NewMonitoringService("test", "./logs")
	if err != nil {
		fmt.Printf("Error creating monitoring service: %v\n", err)
		os.Exit(1)
	}
	
	// Start monitoring
	service.Start()
	
	// Create and record metrics
	counter := service.CreateCounter("test_counter", "Test counter metric")
	gauge := service.CreateGauge("test_gauge", "Test gauge metric")
	histogram := service.CreateHistogram("test_histogram", "Test histogram metric", []float64{0.1, 0.5, 1.0, 5.0})
	
	// Record some values
	counter.Inc()
	gauge.Set(42.0)
	histogram.Observe(0.7)
	
	// Create and trigger an alert
	alertRule := monitoring.AlertRule{
		Name:        "test_alert",
		Description: "Test alert rule",
		Severity:    "warning",
		Condition:   "test_gauge > 40",
	}
	
	service.RegisterAlertRule(alertRule)
	
	// Wait for metrics to be processed
	time.Sleep(1 * time.Second)
	
	// Get and print metrics
	metrics := service.GetMetrics()
	fmt.Println("Metrics:")
	for name, value := range metrics {
		fmt.Printf("  %s: %v\n", name, value)
	}
	
	// Stop monitoring
	service.Stop()
	
	fmt.Println("Monitoring test completed successfully")
}
EOF
        
        # Build the component
        cd "$TEMP_DIR"
        go build -o "$OLDPWD/$BUILD_DIR/monitoring_test" ./src/monitoring_test.go
        
        echo "Monitoring component built successfully: $BUILD_DIR/monitoring_test"
        ;;
        
    *)
        echo "Unknown component: $COMPONENT"
        echo "Available components:"
        echo "  template_security - Template security verification tool"
        echo "  audit_logger - Audit logging functionality"
        echo "  memory_optimizer - Memory optimization components"
        echo "  monitoring - Monitoring and metrics components"
        exit 1
        ;;
esac

echo "Build completed successfully"
