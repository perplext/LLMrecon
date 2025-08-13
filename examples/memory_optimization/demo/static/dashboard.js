// Dashboard.js - Enhanced monitoring dashboard for the static file handler

// Initialize the dashboard when the DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Initialize charts
    initializeCharts();
    
    // Set up refresh interval
    setInterval(refreshDashboard, 5000);
    
    // Initial refresh
    refreshDashboard();
    
    // Set up event listeners
    document.getElementById('time-range').addEventListener('change', refreshDashboard);
    document.getElementById('metric-filter').addEventListener('change', refreshDashboard);
    document.getElementById('refresh-button').addEventListener('click', refreshDashboard);
});

// Global chart objects
let cacheChart, serveTimeChart, compressionChart, alertsChart, memoryChart, requestsChart;

// Initialize charts with empty data
function initializeCharts() {
    // Memory usage chart
    const memoryCtx = document.getElementById('memory-chart').getContext('2d');
    memoryChart = new Chart(memoryCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Heap Allocation (MB)',
                data: [],
                borderColor: '#36a2eb',
                backgroundColor: 'rgba(54, 162, 235, 0.2)',
                tension: 0.1,
                fill: true
            }, {
                label: 'Heap Objects (thousands)',
                data: [],
                borderColor: '#ff6384',
                backgroundColor: 'rgba(255, 99, 132, 0.1)',
                tension: 0.1,
                fill: true,
                yAxisID: 'y1'
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Memory (MB)'
                    }
                },
                y1: {
                    beginAtZero: true,
                    position: 'right',
                    grid: {
                        drawOnChartArea: false
                    },
                    title: {
                        display: true,
                        text: 'Objects (thousands)'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Time'
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Memory Usage'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            let label = context.dataset.label || '';
                            if (label) {
                                label += ': ';
                            }
                            if (context.parsed.y !== null) {
                                if (context.datasetIndex === 0) {
                                    label += context.parsed.y.toFixed(2) + ' MB';
                                } else {
                                    label += context.parsed.y.toFixed(2) + 'K objects';
                                }
                            }
                            return label;
                        }
                    }
                }
            }
        }
    });
    
    // Requests chart
    const requestsCtx = document.getElementById('requests-chart').getContext('2d');
    requestsChart = new Chart(requestsCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Requests per Second',
                data: [],
                borderColor: '#4bc0c0',
                backgroundColor: 'rgba(75, 192, 192, 0.2)',
                tension: 0.1,
                fill: true
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Requests/s'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Time'
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Request Rate'
                }
            }
        }
    });
    // Cache performance chart
    const cacheCtx = document.getElementById('cache-chart').getContext('2d');
    cacheChart = new Chart(cacheCtx, {
        type: 'doughnut',
        data: {
            labels: ['Cache Hits', 'Cache Misses'],
            datasets: [{
                data: [0, 0],
                backgroundColor: ['#36a2eb', '#ff6384']
            }]
        },
        options: {
            responsive: true,
            plugins: {
                title: {
                    display: true,
                    text: 'Cache Performance'
                },
                legend: {
                    position: 'bottom'
                }
            }
        }
    });
    
    // Serve time chart
    const serveTimeCtx = document.getElementById('serve-time-chart').getContext('2d');
    serveTimeChart = new Chart(serveTimeCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Average Serve Time (ms)',
                data: [],
                borderColor: '#4bc0c0',
                tension: 0.1,
                fill: false
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Milliseconds'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Time'
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Average Serve Time'
                }
            }
        }
    });
    
    // Compression chart
    const compressionCtx = document.getElementById('compression-chart').getContext('2d');
    compressionChart = new Chart(compressionCtx, {
        type: 'bar',
        data: {
            labels: ['Original Size', 'Compressed Size'],
            datasets: [{
                label: 'Size (MB)',
                data: [0, 0],
                backgroundColor: ['#ff9f40', '#9966ff']
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Size (MB)'
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Compression Performance'
                }
            }
        }
    });
    
    // Alerts chart
    const alertsCtx = document.getElementById('alerts-chart').getContext('2d');
    alertsChart = new Chart(alertsCtx, {
        type: 'bar',
        data: {
            labels: ['Info', 'Warning', 'Error'],
            datasets: [{
                label: 'Count',
                data: [0, 0, 0],
                backgroundColor: ['#36a2eb', '#ffcd56', '#ff6384']
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        stepSize: 1
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Active Alerts'
                }
            }
        }
    });
}

// Refresh the dashboard with the latest metrics
function refreshDashboard() {
    // Get selected time range
    const timeRange = document.getElementById('time-range').value || 'hour';
    
    // Get selected metric filter
    const metricFilter = document.getElementById('metric-filter').value || 'all';
    
    // Fetch the latest metrics
    fetch('/stats?range=' + timeRange + '&filter=' + metricFilter)
        .then(response => response.json())
        .then(data => {
            // Update summary metrics
            updateSummaryMetrics(data);
            
            // Update charts
            updateCharts(data);
            
            // Update alerts
            updateAlerts(data);
            
            // Update performance indicators
            updatePerformanceIndicators(data);
            
            // Update last updated time
            document.getElementById('last-updated').textContent = new Date().toLocaleTimeString();
            
            // Hide error message if it was previously shown
            document.getElementById('error-container').style.display = 'none';
        })
        .catch(error => {
            console.error('Error fetching metrics:', error);
            document.getElementById('error-message').textContent = 'Error fetching metrics: ' + error.message;
            document.getElementById('error-container').style.display = 'block';
        });
}

// Update summary metrics
function updateSummaryMetrics(data) {
    // Update request count
    document.getElementById('request-count').textContent = data.filesServed.toLocaleString();
    
    // Update cache hit ratio
    const cacheHits = data.cacheHits || 0;
    const cacheMisses = data.cacheMisses || 0;
    const cacheHitRatio = (cacheHits / (cacheHits + cacheMisses)) * 100 || 0;
    document.getElementById('cache-hit-ratio').textContent = cacheHitRatio.toFixed(2) + '%';
    
    // Update average serve time
    document.getElementById('average-serve-time').textContent = (data.averageServeTime || 0).toFixed(2) + ' ms';
    
    // Update compression ratio
    const totalSize = data.totalSize || 1;
    const compressedSize = data.compressedSize || 0;
    const compressionRatio = (1 - (compressedSize / totalSize)) * 100 || 0;
    document.getElementById('compression-ratio').textContent = compressionRatio.toFixed(2) + '%';
    
    // Update memory usage
    document.getElementById('memory-usage').textContent = formatBytes(data.heapAlloc || 0);
    
    // Update heap objects
    document.getElementById('heap-objects').textContent = (data.heapObjects || 0).toLocaleString();
    
    // Update GC CPU fraction
    const gcCPUFraction = (data.gcCPUFraction || 0) * 100;
    document.getElementById('gc-cpu').textContent = gcCPUFraction.toFixed(2) + '%';
    
    // Update alert count
    document.getElementById('alert-count').textContent = (data.alerts || []).length;
}

// Update charts with the latest data
function updateCharts(data) {
    // Update cache chart
    cacheChart.data.datasets[0].data = [data.cacheHits || 0, data.cacheMisses || 0];
    cacheChart.update();
    
    // Update serve time chart
    if (data.serveTimeHistory && data.serveTimeHistory.length > 0) {
        serveTimeChart.data.labels = data.serveTimeHistory.map(item => item.time);
        serveTimeChart.data.datasets[0].data = data.serveTimeHistory.map(item => item.value);
        serveTimeChart.update();
    }
    
    // Update compression chart
    const totalSizeMB = (data.totalSize || 0) / (1024 * 1024);
    const compressedSizeMB = (data.compressedSize || 0) / (1024 * 1024);
    compressionChart.data.datasets[0].data = [totalSizeMB, compressedSizeMB];
    compressionChart.update();
    
    // Update memory chart
    if (data.memoryHistory && data.memoryHistory.length > 0) {
        memoryChart.data.labels = data.memoryHistory.map(item => item.time);
        memoryChart.data.datasets[0].data = data.memoryHistory.map(item => item.heapAlloc / (1024 * 1024));
        memoryChart.data.datasets[1].data = data.memoryHistory.map(item => item.heapObjects / 1000);
        memoryChart.update();
    }
    
    // Update requests chart
    if (data.requestRateHistory && data.requestRateHistory.length > 0) {
        requestsChart.data.labels = data.requestRateHistory.map(item => item.time);
        requestsChart.data.datasets[0].data = data.requestRateHistory.map(item => item.value);
        requestsChart.update();
    }
    
    // Update alerts chart if there's history data
    if (data.alertHistory && data.alertHistory.length > 0) {
        alertsChart.data.labels = data.alertHistory.map(item => item.time);
        alertsChart.data.datasets[0].data = data.alertHistory.map(item => item.count);
        alertsChart.update();
    }
}

// Update alerts section
function updateAlerts(data) {
    const alertsContainer = document.getElementById('alerts-container');
    alertsContainer.innerHTML = '';
    
    // Simulate some alerts based on the metrics
    const alerts = [];
    
    // Check cache hit ratio
    if (data.monitoring.cacheHitRatio < 0.5) {
        alerts.push({
            severity: 'warning',
            message: 'Low cache hit ratio: ' + (data.monitoring.cacheHitRatio * 100).toFixed(2) + '%',
            time: new Date().toLocaleTimeString()
        });
    }
    
    // Check average serve time
    if (data.monitoring.averageServeTimeMs > 50) {
        alerts.push({
            severity: 'warning',
            message: 'Slow average serve time: ' + data.monitoring.averageServeTimeMs + ' ms',
            time: new Date().toLocaleTimeString()
        });
    }
    
    // Check cache size (assume max cache size is 100MB)
    const maxCacheSize = 100 * 1024 * 1024;
    if (data.monitoring.cacheSize > 0.9 * maxCacheSize) {
        alerts.push({
            severity: 'warning',
            message: 'Cache nearly full: ' + formatBytes(data.monitoring.cacheSize) + ' / ' + formatBytes(maxCacheSize),
            time: new Date().toLocaleTimeString()
        });
    }
    
    // Add info alert for demonstration
    if (alerts.length === 0) {
        alerts.push({
            severity: 'info',
            message: 'All metrics within normal ranges',
            time: new Date().toLocaleTimeString()
        });
    }
    
    // Add alerts to the container
    alerts.forEach(alert => {
        const alertElement = document.createElement('div');
        alertElement.className = 'alert alert-' + alert.severity;
        alertElement.innerHTML = `
            <div class="alert-time">${alert.time}</div>
            <div class="alert-message">${alert.message}</div>
        `;
        alertsContainer.appendChild(alertElement);
    });
    
    // Update alert count
    document.getElementById('alert-count').textContent = alerts.length;
}

// Format bytes to human-readable format
function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Update performance indicators
function updatePerformanceIndicators(data) {
    // Calculate performance indicators
    const cacheHitRatio = (data.cacheHits / (data.cacheHits + data.cacheMisses)) * 100 || 0;
    const averageServeTime = data.averageServeTime || 0;
    const memoryUsage = data.heapAlloc || 0;
    const requestRate = data.requestRate || 0;
    
    // Update performance status indicators
    updatePerformanceStatus('cache-status', cacheHitRatio, 70, 50);
    updatePerformanceStatus('time-status', averageServeTime, 10, 50, true);
    updatePerformanceStatus('memory-status', memoryUsage / (1024 * 1024), 100, 500, true);
    updatePerformanceStatus('request-status', requestRate, 50, 10);
    
    // Update optimization metrics
    if (data.optimizationMetrics) {
        const memoryReduction = data.optimizationMetrics.memoryReduction || 0;
        const throughputIncrease = data.optimizationMetrics.throughputIncrease || 0;
        
        document.getElementById('memory-reduction').textContent = memoryReduction.toFixed(2) + '%';
        document.getElementById('throughput-increase').textContent = throughputIncrease.toFixed(2) + '%';
        
        // Highlight if we've met our optimization targets
        if (memoryReduction >= 25) {
            document.getElementById('memory-reduction').classList.add('target-met');
        } else {
            document.getElementById('memory-reduction').classList.remove('target-met');
        }
        
        if (throughputIncrease >= 100) { // 2x = 100% increase
            document.getElementById('throughput-increase').classList.add('target-met');
        } else {
            document.getElementById('throughput-increase').classList.remove('target-met');
        }
    }
}

// Update performance status indicator
function updatePerformanceStatus(elementId, value, goodThreshold, badThreshold, inversed = false) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    // Remove existing classes
    element.classList.remove('status-good', 'status-warning', 'status-bad');
    
    // Determine status based on thresholds
    let status;
    if (inversed) {
        // For metrics where lower is better (like serve time)
        if (value <= goodThreshold) {
            status = 'status-good';
        } else if (value <= badThreshold) {
            status = 'status-warning';
        } else {
            status = 'status-bad';
        }
    } else {
        // For metrics where higher is better (like cache hit ratio)
        if (value >= goodThreshold) {
            status = 'status-good';
        } else if (value >= badThreshold) {
            status = 'status-warning';
        } else {
            status = 'status-bad';
        }
    }
    
    // Apply status class
    element.classList.add(status);
}
