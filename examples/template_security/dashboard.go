package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/perplext/LLMrecon/src/template/security/sandbox"
)

// DashboardServer represents a server for the security dashboard
type DashboardServer struct {
	// framework is the security framework
	framework *sandbox.SecurityFramework
	// port is the server port
	port int
	// templateDir is the directory for HTML templates
	templateDir string
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(framework *sandbox.SecurityFramework, port int, templateDir string) *DashboardServer {
	return &DashboardServer{
		framework:   framework,
		port:        port,
		templateDir: templateDir,
	}
}

// Start starts the dashboard server
func (s *DashboardServer) Start() error {
	// Create the template directory if it doesn't exist
	if err := os.MkdirAll(s.templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Create the HTML templates
	if err := s.createTemplates(); err != nil {
		return fmt.Errorf("failed to create templates: %w", err)
	}

	// Set up the HTTP handlers
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/metrics", s.handleMetrics)
	http.HandleFunc("/api/alerts", s.handleAlerts)
	http.HandleFunc("/api/clear-alerts", s.handleClearAlerts)
	http.HandleFunc("/api/reset-metrics", s.handleResetMetrics)

	// Start the server
	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Starting dashboard server on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// createTemplates creates the HTML templates
func (s *DashboardServer) createTemplates() error {
	// Create the index template
	indexTemplate := filepath.Join(s.templateDir, "index.html")
	if err := os.WriteFile(indexTemplate, []byte(indexHTML), 0644); err != nil {
		return fmt.Errorf("failed to create index template: %w", err)
	}

	return nil
}

// handleIndex handles the index page
func (s *DashboardServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Parse the template
	tmpl, err := template.ParseFiles(filepath.Join(s.templateDir, "index.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	// Execute the template
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleMetrics handles the metrics API
func (s *DashboardServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Get the metrics
	metrics := s.framework.GetMetrics()

	// Return the metrics as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode metrics: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleAlerts handles the alerts API
func (s *DashboardServer) handleAlerts(w http.ResponseWriter, r *http.Request) {
	// Get the alerts
	alerts := s.framework.GetAlerts()

	// Return the alerts as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode alerts: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleClearAlerts handles the clear alerts API
func (s *DashboardServer) handleClearAlerts(w http.ResponseWriter, r *http.Request) {
	// Clear the alerts
	s.framework.ClearAlerts()

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

// handleResetMetrics handles the reset metrics API
func (s *DashboardServer) handleResetMetrics(w http.ResponseWriter, r *http.Request) {
	// Reset the metrics
	s.framework.ResetMetrics()

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

// RunDashboard runs the security dashboard
func RunDashboard(framework *sandbox.SecurityFramework, port int) {
	// Create the template directory
	templateDir := "./dashboard_templates"

	// Create the dashboard server
	server := NewDashboardServer(framework, port, templateDir)

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start dashboard server: %v", err)
	}
}

// indexHTML is the HTML template for the index page
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Template Security Dashboard</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background-color: #333;
            color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .card {
            background-color: white;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            padding: 20px;
            margin-bottom: 20px;
        }
        .card-title {
            margin-top: 0;
            border-bottom: 1px solid #eee;
            padding-bottom: 10px;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
        }
        .metric {
            background-color: #f9f9f9;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #333;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            margin: 10px 0;
        }
        .metric-label {
            color: #666;
            font-size: 14px;
        }
        .alert {
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 5px;
        }
        .alert-info {
            background-color: #d1ecf1;
            border-left: 4px solid #0c5460;
        }
        .alert-warning {
            background-color: #fff3cd;
            border-left: 4px solid #856404;
        }
        .alert-error {
            background-color: #f8d7da;
            border-left: 4px solid #721c24;
        }
        .alert-critical {
            background-color: #f8d7da;
            border-left: 4px solid #721c24;
            font-weight: bold;
        }
        .alert-time {
            font-size: 12px;
            color: #666;
        }
        .alert-message {
            margin: 5px 0;
        }
        .alert-template {
            font-style: italic;
            font-size: 14px;
        }
        .button {
            background-color: #333;
            color: white;
            border: none;
            padding: 10px 15px;
            border-radius: 5px;
            cursor: pointer;
            margin-right: 10px;
        }
        .button:hover {
            background-color: #555;
        }
        .refresh-container {
            margin-bottom: 20px;
            text-align: right;
        }
        .chart-container {
            height: 300px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Template Security Dashboard</h1>
            <p>Real-time monitoring of template validation, execution, and workflow metrics</p>
        </div>

        <div class="refresh-container">
            <button class="button" id="refresh-button">Refresh Data</button>
            <button class="button" id="clear-alerts-button">Clear Alerts</button>
            <button class="button" id="reset-metrics-button">Reset Metrics</button>
        </div>

        <div class="card">
            <h2 class="card-title">Validation Metrics</h2>
            <div class="metrics-grid" id="validation-metrics">
                <div class="metric">
                    <div class="metric-label">Templates Validated</div>
                    <div class="metric-value" id="validation-count">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Validation Errors</div>
                    <div class="metric-value" id="validation-errors">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Average Validation Time</div>
                    <div class="metric-value" id="avg-validation-time">0ms</div>
                </div>
            </div>
            <div class="chart-container">
                <canvas id="risk-chart"></canvas>
            </div>
        </div>

        <div class="card">
            <h2 class="card-title">Execution Metrics</h2>
            <div class="metrics-grid" id="execution-metrics">
                <div class="metric">
                    <div class="metric-label">Templates Executed</div>
                    <div class="metric-value" id="execution-count">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Execution Errors</div>
                    <div class="metric-value" id="execution-errors">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Average Execution Time</div>
                    <div class="metric-value" id="avg-execution-time">0ms</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Average CPU Time</div>
                    <div class="metric-value" id="avg-cpu-time">0s</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Average Memory Usage</div>
                    <div class="metric-value" id="avg-memory-usage">0MB</div>
                </div>
            </div>
            <div class="chart-container">
                <canvas id="execution-chart"></canvas>
            </div>
        </div>

        <div class="card">
            <h2 class="card-title">Workflow Metrics</h2>
            <div class="metrics-grid" id="workflow-metrics">
                <div class="metric">
                    <div class="metric-label">Template Versions</div>
                    <div class="metric-value" id="template-versions">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Approved Templates</div>
                    <div class="metric-value" id="approved-templates">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Rejected Templates</div>
                    <div class="metric-value" id="rejected-templates">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Pending Templates</div>
                    <div class="metric-value" id="pending-templates">0</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Deprecated Templates</div>
                    <div class="metric-value" id="deprecated-templates">0</div>
                </div>
            </div>
            <div class="chart-container">
                <canvas id="workflow-chart"></canvas>
            </div>
        </div>

        <div class="card">
            <h2 class="card-title">Security Alerts</h2>
            <div id="alerts-container">
                <p>No alerts to display.</p>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script>
        // Charts
        let riskChart = null;
        let executionChart = null;
        let workflowChart = null;

        // Initialize the dashboard
        document.addEventListener('DOMContentLoaded', function() {
            // Load the initial data
            loadData();

            // Set up the refresh button
            document.getElementById('refresh-button').addEventListener('click', loadData);

            // Set up the clear alerts button
            document.getElementById('clear-alerts-button').addEventListener('click', clearAlerts);

            // Set up the reset metrics button
            document.getElementById('reset-metrics-button').addEventListener('click', resetMetrics);

            // Set up automatic refresh every 30 seconds
            setInterval(loadData, 30000);
        });

        // Load data from the API
        function loadData() {
            // Load metrics
            fetch('/api/metrics')
                .then(response => response.json())
                .then(data => {
                    updateMetrics(data);
                    updateCharts(data);
                })
                .catch(error => console.error('Error loading metrics:', error));

            // Load alerts
            fetch('/api/alerts')
                .then(response => response.json())
                .then(data => {
                    updateAlerts(data);
                })
                .catch(error => console.error('Error loading alerts:', error));
        }

        // Update metrics display
        function updateMetrics(data) {
            // Check if metrics are enabled
            if (data.metrics_enabled === false) {
                document.getElementById('validation-count').textContent = 'N/A';
                document.getElementById('validation-errors').textContent = 'N/A';
                document.getElementById('avg-validation-time').textContent = 'N/A';
                document.getElementById('execution-count').textContent = 'N/A';
                document.getElementById('execution-errors').textContent = 'N/A';
                document.getElementById('avg-execution-time').textContent = 'N/A';
                document.getElementById('avg-cpu-time').textContent = 'N/A';
                document.getElementById('avg-memory-usage').textContent = 'N/A';
                document.getElementById('template-versions').textContent = 'N/A';
                document.getElementById('approved-templates').textContent = 'N/A';
                document.getElementById('rejected-templates').textContent = 'N/A';
                document.getElementById('pending-templates').textContent = 'N/A';
                document.getElementById('deprecated-templates').textContent = 'N/A';
                return;
            }

            // Update validation metrics
            if (data.validation) {
                document.getElementById('validation-count').textContent = data.validation.validationCount || 0;
                document.getElementById('validation-errors').textContent = data.validation.validationErrors || 0;
                document.getElementById('avg-validation-time').textContent = data.validation.averageValidationTime || '0ms';
            }

            // Update execution metrics
            if (data.execution) {
                document.getElementById('execution-count').textContent = data.execution.executionCount || 0;
                document.getElementById('execution-errors').textContent = data.execution.executionErrors || 0;
                document.getElementById('avg-execution-time').textContent = data.execution.averageExecutionTime || '0ms';
                
                if (data.execution.averageResourceUsage) {
                    document.getElementById('avg-cpu-time').textContent = (data.execution.averageResourceUsage.cpuTime || 0).toFixed(2) + 's';
                    document.getElementById('avg-memory-usage').textContent = (data.execution.averageResourceUsage.memoryUsage || 0) + 'MB';
                }
            }

            // Update workflow metrics
            if (data.workflow) {
                document.getElementById('template-versions').textContent = data.workflow.templateVersions || 0;
                document.getElementById('approved-templates').textContent = data.workflow.approvedTemplates || 0;
                document.getElementById('rejected-templates').textContent = data.workflow.rejectedTemplates || 0;
                document.getElementById('pending-templates').textContent = data.workflow.pendingTemplates || 0;
                document.getElementById('deprecated-templates').textContent = data.workflow.deprecatedTemplates || 0;
            }
        }

        // Update charts
        function updateCharts(data) {
            // Check if metrics are enabled
            if (data.metrics_enabled === false) {
                return;
            }

            // Update risk chart
            if (data.validation && data.validation.templatesByRisk) {
                const riskData = {
                    labels: Object.keys(data.validation.templatesByRisk),
                    datasets: [{
                        label: 'Templates by Risk Category',
                        data: Object.values(data.validation.templatesByRisk),
                        backgroundColor: [
                            'rgba(75, 192, 192, 0.2)',
                            'rgba(255, 206, 86, 0.2)',
                            'rgba(255, 159, 64, 0.2)',
                            'rgba(255, 99, 132, 0.2)'
                        ],
                        borderColor: [
                            'rgba(75, 192, 192, 1)',
                            'rgba(255, 206, 86, 1)',
                            'rgba(255, 159, 64, 1)',
                            'rgba(255, 99, 132, 1)'
                        ],
                        borderWidth: 1
                    }]
                };

                if (riskChart) {
                    riskChart.data = riskData;
                    riskChart.update();
                } else {
                    const ctx = document.getElementById('risk-chart').getContext('2d');
                    riskChart = new Chart(ctx, {
                        type: 'bar',
                        data: riskData,
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            scales: {
                                y: {
                                    beginAtZero: true
                                }
                            }
                        }
                    });
                }
            }

            // Update execution chart
            if (data.execution) {
                const executionData = {
                    labels: ['Successful', 'Failed'],
                    datasets: [{
                        label: 'Template Executions',
                        data: [
                            (data.execution.executionCount || 0) - (data.execution.executionErrors || 0),
                            data.execution.executionErrors || 0
                        ],
                        backgroundColor: [
                            'rgba(75, 192, 192, 0.2)',
                            'rgba(255, 99, 132, 0.2)'
                        ],
                        borderColor: [
                            'rgba(75, 192, 192, 1)',
                            'rgba(255, 99, 132, 1)'
                        ],
                        borderWidth: 1
                    }]
                };

                if (executionChart) {
                    executionChart.data = executionData;
                    executionChart.update();
                } else {
                    const ctx = document.getElementById('execution-chart').getContext('2d');
                    executionChart = new Chart(ctx, {
                        type: 'pie',
                        data: executionData,
                        options: {
                            responsive: true,
                            maintainAspectRatio: false
                        }
                    });
                }
            }

            // Update workflow chart
            if (data.workflow) {
                const workflowData = {
                    labels: ['Approved', 'Rejected', 'Pending', 'Deprecated'],
                    datasets: [{
                        label: 'Template Workflow Status',
                        data: [
                            data.workflow.approvedTemplates || 0,
                            data.workflow.rejectedTemplates || 0,
                            data.workflow.pendingTemplates || 0,
                            data.workflow.deprecatedTemplates || 0
                        ],
                        backgroundColor: [
                            'rgba(75, 192, 192, 0.2)',
                            'rgba(255, 99, 132, 0.2)',
                            'rgba(255, 206, 86, 0.2)',
                            'rgba(153, 102, 255, 0.2)'
                        ],
                        borderColor: [
                            'rgba(75, 192, 192, 1)',
                            'rgba(255, 99, 132, 1)',
                            'rgba(255, 206, 86, 1)',
                            'rgba(153, 102, 255, 1)'
                        ],
                        borderWidth: 1
                    }]
                };

                if (workflowChart) {
                    workflowChart.data = workflowData;
                    workflowChart.update();
                } else {
                    const ctx = document.getElementById('workflow-chart').getContext('2d');
                    workflowChart = new Chart(ctx, {
                        type: 'doughnut',
                        data: workflowData,
                        options: {
                            responsive: true,
                            maintainAspectRatio: false
                        }
                    });
                }
            }
        }

        // Update alerts display
        function updateAlerts(alerts) {
            const container = document.getElementById('alerts-container');
            
            if (!alerts || alerts.length === 0) {
                container.innerHTML = '<p>No alerts to display.</p>';
                return;
            }

            // Sort alerts by timestamp (newest first)
            alerts.sort((a, b) => new Date(b.Timestamp) - new Date(a.Timestamp));

            // Clear the container
            container.innerHTML = '';

            // Add the alerts
            alerts.forEach(alert => {
                const alertDiv = document.createElement('div');
                alertDiv.className = 'alert alert-' + alert.Level.toLowerCase();

                const time = new Date(alert.Timestamp).toLocaleString();
                alertDiv.innerHTML = `
                    <div class="alert-time">${time}</div>
                    <div class="alert-message">${alert.Message}</div>
                    <div class="alert-template">Template: ${alert.TemplateID}</div>
                `;

                container.appendChild(alertDiv);
            });
        }

        // Clear alerts
        function clearAlerts() {
            fetch('/api/clear-alerts', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        document.getElementById('alerts-container').innerHTML = '<p>No alerts to display.</p>';
                    }
                })
                .catch(error => console.error('Error clearing alerts:', error));
        }

        // Reset metrics
        function resetMetrics() {
            fetch('/api/reset-metrics', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        loadData();
                    }
                })
                .catch(error => console.error('Error resetting metrics:', error));
        }
    </script>
</body>
</html>
`
