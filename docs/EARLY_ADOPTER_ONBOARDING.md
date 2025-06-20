# LLM Red Team v0.2.0 Early Adopter Onboarding

Welcome to the LLM Red Team v0.2.0 Early Adopter Program! This guide will help you deploy and optimize the distributed infrastructure for your organization.

## 🚀 Quick Start Checklist

- [ ] Review system requirements
- [ ] Deploy infrastructure components
- [ ] Configure monitoring
- [ ] Run validation tests
- [ ] Schedule training session
- [ ] Join Slack channel

## 📋 System Requirements

### Minimum Requirements
- **CPU**: 8 cores
- **Memory**: 16GB RAM
- **Storage**: 100GB SSD
- **OS**: Linux (Ubuntu 20.04+) or macOS
- **Go**: 1.21+
- **Redis**: 6.2+ (cluster mode recommended)

### Recommended Production Setup
- **CPU**: 16+ cores
- **Memory**: 32GB RAM
- **Storage**: 500GB SSD with high IOPS
- **Network**: 1Gbps+ connection
- **Redis Cluster**: 6 nodes minimum

## 🔧 Deployment Steps

### 1. Infrastructure Setup

```bash
# Clone repository
git clone https://github.com/your-org/llm-red-team.git
cd llm-red-team

# Checkout v0.2.0
git checkout v0.2.0

# Install dependencies
go mod download

# Build application
go build -o llm-red-team ./src/main.go
```

### 2. Redis Cluster Configuration

```bash
# For development (single instance)
redis-server --port 6379

# For production (cluster mode)
# See redis-cluster-config.conf in deployment/redis/
```

### 3. Application Configuration

Create `config.yaml`:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: distributed

redis:
  cluster:
    nodes:
      - localhost:7000
      - localhost:7001
      - localhost:7002
    password: ""
    
performance:
  max_concurrent_attacks: 100
  worker_pool_size: 50
  queue_buffer_size: 1000
  
monitoring:
  enabled: true
  port: 8090
  
providers:
  - name: openai
    api_key: ${OPENAI_API_KEY}
    rate_limit: 100
  - name: anthropic
    api_key: ${ANTHROPIC_API_KEY}
    rate_limit: 50
```

### 4. Start Services

```bash
# Start with distributed mode
./llm-red-team server --config config.yaml --distributed

# Verify health
curl http://localhost:8080/health
curl http://localhost:8090/api/v1/status
```

## 📊 Performance Tuning

### Initial Benchmarking

Run performance validation:

```bash
./scripts/validate_production.sh
./scripts/quick_perf_test.sh 100 1000
```

Expected results:
- Success rate: >95%
- Response time: <2s
- Throughput: >100 req/s

### Optimization Guidelines

1. **Connection Pooling**
   ```yaml
   redis:
     pool_size: 100
     min_idle_conns: 10
     max_conn_age: 30m
   ```

2. **Worker Tuning**
   ```yaml
   performance:
     worker_pool_size: runtime.NumCPU() * 10
     queue_workers: runtime.NumCPU() * 5
   ```

3. **Memory Management**
   ```yaml
   runtime:
     gc_percent: 100
     max_memory: 8GB
   ```

## 🔍 Monitoring Setup

### 1. Prometheus Integration

```bash
# Copy configurations
cp monitoring/prometheus-config.yml /etc/prometheus/
cp monitoring/alerts.yml /etc/prometheus/

# Start Prometheus
prometheus --config.file=/etc/prometheus/prometheus.yml
```

### 2. Grafana Dashboard

1. Access Grafana: http://localhost:3000
2. Import dashboard: `monitoring/grafana-dashboard.json`
3. Configure alerts to your notification channels

### 3. Key Metrics to Watch

- **Performance**: Response time, success rate, throughput
- **Resources**: CPU, memory, goroutines
- **Queue**: Depth, processing rate
- **Errors**: Error rate, types

## 🎯 Use Case Examples

### 1. Automated Security Testing

```python
import requests

# Configure attack campaign
campaign = {
    "name": "Q1 Security Audit",
    "targets": ["gpt-4", "claude-2", "gemini-pro"],
    "attack_types": ["prompt_injection", "jailbreak"],
    "concurrent": 50,
    "duration": "1h"
}

# Launch campaign
response = requests.post(
    "http://localhost:8080/api/v1/campaigns",
    json=campaign
)
```

### 2. Continuous Monitoring

```bash
# Schedule daily security scans
0 2 * * * /opt/llm-red-team/scripts/daily_security_scan.sh

# Real-time alerting
curl -X POST http://localhost:8080/api/v1/alerts \
  -d '{"webhook": "https://slack.com/your-webhook"}'
```

### 3. Compliance Reporting

```bash
# Generate OWASP compliance report
./llm-red-team report owasp --format pdf --output reports/

# ISO 42001 assessment
./llm-red-team compliance iso42001 --detailed
```

## 🐛 Troubleshooting Guide

### Common Issues

1. **High Response Times**
   - Check Redis latency: `redis-cli --latency`
   - Review connection pool settings
   - Verify network connectivity

2. **Memory Growth**
   - Enable profiling: `http://localhost:6060/debug/pprof/`
   - Check for goroutine leaks
   - Review object pooling

3. **Failed Attacks**
   - Verify API credentials
   - Check rate limits
   - Review provider status

### Debug Commands

```bash
# Check system health
curl http://localhost:8080/api/v1/health/detailed

# View current metrics
curl http://localhost:8090/api/v1/metrics

# Export debug bundle
./llm-red-team debug export --output debug-bundle.zip
```

## 📞 Support Channels

### Immediate Support
- **Slack**: #llm-red-team-early-adopters
- **Emergency**: security-team@company.com
- **Office Hours**: Tue/Thu 2-3 PM PST

### Resources
- **Documentation**: /docs
- **API Reference**: http://localhost:8080/swagger
- **Video Tutorials**: Available on request

### Feedback Collection
We value your feedback! Please share:
- Performance observations
- Feature requests
- Bug reports
- Use case scenarios

Submit via:
- GitHub Issues
- Slack channel
- Monthly survey

## 🗓️ Early Adopter Timeline

### Week 1-2: Deployment & Testing
- Deploy infrastructure
- Run initial benchmarks
- Configure monitoring

### Week 3-4: Production Usage
- Launch security campaigns
- Monitor performance
- Collect metrics

### Week 5-6: Optimization
- Tune performance
- Implement feedback
- Share learnings

## 🎉 Success Criteria

Your deployment is successful when:
- [ ] 100+ concurrent attacks sustained
- [ ] <2s average response time
- [ ] >95% success rate
- [ ] Monitoring operational
- [ ] Team trained

## 📝 Feedback Form

Please complete after 2 weeks:
https://forms.gle/llm-red-team-v020-feedback

---

**Thank you for being an early adopter! Your feedback shapes the future of LLM Red Team.**

*Last Updated: 2025-06-20*
*Version: v0.2.0*