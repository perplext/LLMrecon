# v0.2.0 Monitoring & Alerting Setup Guide

This guide helps early v0.2.0 adopters set up comprehensive monitoring and alerting for their LLM Red Team deployments.

## Quick Start

```bash
# 1. Install monitoring stack (if not already installed)
brew install prometheus grafana

# 2. Copy monitoring configurations
cp monitoring/prometheus-config.yml /usr/local/etc/prometheus/prometheus.yml
cp monitoring/alerts.yml /usr/local/etc/prometheus/

# 3. Start services
brew services start prometheus
brew services start grafana

# 4. Import dashboard
# Open http://localhost:3000 (admin/admin)
# Import monitoring/grafana-dashboard.json
```

## Architecture Overview

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  LLM Red Team   │────▶│  Prometheus  │────▶│   Grafana   │
│   Metrics API   │     │   Scraper    │     │  Dashboard  │
└─────────────────┘     └──────────────┘     └─────────────┘
         │                      │                     │
         │                      ▼                     ▼
         │              ┌──────────────┐     ┌─────────────┐
         └─────────────▶│ Alert Manager│────▶│ PagerDuty/  │
                        │              │     │   Slack     │
                        └──────────────┘     └─────────────┘
```

## Key Metrics to Monitor

### 1. Performance Metrics
- **Concurrent Attacks**: Current number of concurrent attacks
- **Success Rate**: Percentage of successful attacks (target: ≥95%)
- **Response Time**: Average response time in seconds (target: <2s)
- **Throughput**: Requests per second

### 2. Resource Metrics
- **CPU Usage**: Percentage of CPU utilization (alert: >80%)
- **Memory Usage**: Memory consumption (alert: >8GB)
- **Goroutines**: Number of active goroutines
- **Queue Depth**: Pending jobs in queue

### 3. Infrastructure Health
- **Redis Status**: Connection health and memory usage
- **API Availability**: Health check status
- **Error Rate**: Rate of errors per minute

## Alert Configuration

### Critical Alerts (Immediate Action Required)
1. **Application Down**: Main API not responding
2. **Redis Down**: Redis connection lost
3. **Low Success Rate**: Success rate <95% for 5+ minutes

### Warning Alerts (Investigation Needed)
1. **High Response Time**: >2s average response time
2. **High Resource Usage**: CPU >80% or Memory >8GB
3. **Queue Backlog**: >1000 items in queue

## Grafana Dashboard Features

### Main Dashboard (llm-red-team-v020)
- **Real-time Performance**: Live attack metrics
- **Resource Utilization**: CPU, memory, and goroutines
- **Success/Failure Tracking**: Visual success rate gauge
- **Response Time Distribution**: Histogram of response times

### Custom Panels
1. **Attack Types Distribution**: Breakdown by attack type
2. **Provider Performance**: Metrics per LLM provider
3. **Geographic Distribution**: Attack sources (if applicable)
4. **Error Analysis**: Common error patterns

## Integration with Alert Systems

### Slack Integration
```yaml
# alertmanager.yml
receivers:
  - name: 'slack-notifications'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK_URL'
        channel: '#llm-red-team-alerts'
        title: 'LLM Red Team Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

### PagerDuty Integration
```yaml
# alertmanager.yml
receivers:
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
        description: '{{ .GroupLabels.alertname }}'
```

## Monitoring Best Practices

### 1. Baseline Establishment
- Run system for 24-48 hours to establish baselines
- Document normal operating ranges
- Adjust alert thresholds based on actual usage

### 2. Regular Reviews
- Weekly performance reviews
- Monthly capacity planning
- Quarterly alert tuning

### 3. Incident Response
- Document all incidents and resolutions
- Update runbooks based on learnings
- Regular incident response drills

## Troubleshooting Common Issues

### High Response Times
1. Check Redis connection latency
2. Verify CPU/memory resources
3. Review concurrent attack limits
4. Analyze slow queries in logs

### Low Success Rates
1. Check provider API limits
2. Verify network connectivity
3. Review error logs for patterns
4. Check authentication/credentials

### Resource Exhaustion
1. Implement connection pooling
2. Tune garbage collection
3. Review memory leaks
4. Scale horizontally if needed

## Custom Metrics Implementation

### Adding Custom Metrics
```go
// Example: Add custom attack metric
attackDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "llm_red_team_attack_duration_seconds",
        Help: "Attack execution duration in seconds",
        Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
    },
    []string{"attack_type", "provider"},
)
prometheus.MustRegister(attackDuration)
```

### Exposing Metrics Endpoint
```go
// In your main application
http.Handle("/metrics", promhttp.Handler())
http.ListenAndServe(":8090", nil)
```

## Performance Optimization Tips

### 1. Metric Collection
- Use histograms for latency metrics
- Implement counters for throughput
- Add labels sparingly (high cardinality impacts performance)

### 2. Dashboard Optimization
- Use appropriate time ranges
- Implement variable templates
- Cache expensive queries

### 3. Alert Tuning
- Start with conservative thresholds
- Use evaluation periods to reduce noise
- Group related alerts

## Compliance & Reporting

### Monthly Reports
- Average success rate
- Peak concurrent attacks
- Resource utilization trends
- Incident summary

### Audit Trail
- All configuration changes
- Alert acknowledgments
- Performance tuning actions
- Capacity upgrades

## Support Resources

### Documentation
- Prometheus: https://prometheus.io/docs
- Grafana: https://grafana.com/docs
- LLM Red Team: /docs

### Community
- GitHub Issues: Report monitoring problems
- Slack Channel: #llm-red-team-monitoring
- Office Hours: Thursdays 2-3 PM PST

---

*Last Updated: 2025-06-20*
*Version: v0.2.0*