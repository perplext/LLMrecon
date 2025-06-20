# Production Deployment Guide - v0.2.0

This guide covers deploying LLM Red Team v0.2.0 in production environments with distributed execution capabilities.

## Overview

v0.2.0 introduces enterprise-grade infrastructure supporting:
- **100+ concurrent attacks** (vs ~10 in previous versions)
- **Distributed execution** across multiple nodes
- **Redis cluster** for caching and job coordination
- **Real-time monitoring** and performance profiling
- **Auto-scaling** and load balancing

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   LLM Node 1    │    │   LLM Node 2    │    │   LLM Node 3    │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Attack Eng. │ │    │ │ Attack Eng. │ │    │ │ Attack Eng. │ │
│ │ Job Workers │ │    │ │ Job Workers │ │    │ │ Job Workers │ │
│ │ Monitoring  │ │    │ │ Monitoring  │ │    │ │ Monitoring  │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │     Redis Cluster       │
                    │                         │
                    │ ┌─────┐ ┌─────┐ ┌─────┐ │
                    │ │ M:1 │ │ M:2 │ │ M:3 │ │
                    │ │ S:4 │ │ S:5 │ │ S:6 │ │
                    │ └─────┘ └─────┘ └─────┘ │
                    └─────────────────────────┘
```

## Infrastructure Requirements

### Redis Cluster
- **Minimum**: 3 master nodes + 3 replica nodes
- **Memory**: 8GB+ per node
- **Network**: Low latency between nodes (<1ms)
- **Persistence**: RDB + AOF enabled
- **Version**: Redis 6.0+

### Application Nodes
- **CPU**: 8+ cores per node
- **Memory**: 16GB+ per node
- **Storage**: 50GB+ SSD
- **Network**: 1Gbps+ bandwidth
- **OS**: Linux (Ubuntu 20.04+ or CentOS 8+)

### Load Balancer
- **Type**: L7 HTTP/HTTPS load balancer
- **Features**: Health checks, session affinity
- **Examples**: HAProxy, NGINX, AWS ALB, GCP Load Balancer

### Monitoring Infrastructure
- **Metrics**: Prometheus + Grafana
- **Logs**: ELK Stack or Loki + Grafana
- **Alerts**: AlertManager or PagerDuty
- **Tracing**: Jaeger (optional)

## Deployment Steps

### 1. Redis Cluster Setup

#### Using Docker Compose
```yaml
# redis-cluster.yml
version: '3.8'
services:
  redis-1:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes --port 7000
    ports: ["7000:7000", "17000:17000"]
    volumes: ["redis-1-data:/data"]
    
  redis-2:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes --port 7001
    ports: ["7001:7001", "17001:17001"]
    volumes: ["redis-2-data:/data"]
    
  redis-3:
    image: redis:7-alpine
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes --port 7002
    ports: ["7002:7002", "17002:17002"]
    volumes: ["redis-3-data:/data"]

volumes:
  redis-1-data:
  redis-2-data:
  redis-3-data:
```

```bash
# Deploy Redis cluster
docker-compose -f redis-cluster.yml up -d

# Initialize cluster
docker exec -it redis-1 redis-cli --cluster create \
  redis-1:7000 redis-2:7001 redis-3:7002 \
  --cluster-replicas 0 --cluster-yes
```

#### Using Kubernetes
```yaml
# redis-cluster.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-cluster-config
data:
  redis.conf: |
    cluster-enabled yes
    cluster-config-file nodes.conf
    cluster-node-timeout 5000
    appendonly yes
    protected-mode no
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
spec:
  serviceName: redis-cluster
  replicas: 6
  selector:
    matchLabels:
      app: redis-cluster
  template:
    metadata:
      labels:
        app: redis-cluster
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server"]
        args: ["/etc/redis/redis.conf"]
        ports:
        - containerPort: 6379
        - containerPort: 16379
        volumeMounts:
        - name: config
          mountPath: /etc/redis
        - name: data
          mountPath: /data
      volumes:
      - name: config
        configMap:
          name: redis-cluster-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

### 2. Application Deployment

#### Production Configuration
```yaml
# production-config.yaml
distributed:
  enabled: true
  node_id: "${NODE_ID}"
  cluster_name: "llm-prod-cluster"
  
redis:
  cluster:
    nodes:
      - "redis-1:7000"
      - "redis-2:7001"
      - "redis-3:7002"
    password: "${REDIS_PASSWORD}"
    max_redirects: 8
    read_timeout: 3s
    write_timeout: 3s

performance:
  connection_pools:
    enabled: true
    max_connections_per_provider: 100
    health_check_interval: 30s
    
  memory_optimization:
    enabled: true
    gc_tuning: true
    object_pooling: true
    
  rate_limiting:
    enabled: true
    algorithm: "token_bucket"
    default_limit: 1000
    default_window: 60s
    
  monitoring:
    dashboard:
      enabled: true
      host: "0.0.0.0"
      port: 8090
      
    profiling:
      enabled: true
      server_port: 6060
      cpu_profiling: true
      memory_profiling: true
      
  caching:
    redis_cluster:
      enabled: true
      compression_enabled: true
      warming_enabled: true
      invalidation_strategy: "ttl"

security:
  tls:
    enabled: true
    cert_file: "/etc/certs/server.crt"
    key_file: "/etc/certs/server.key"
    
  authentication:
    enabled: true
    method: "jwt"
    secret: "${JWT_SECRET}"

logging:
  level: "info"
  format: "json"
  output: ["stdout", "file"]
  file_path: "/var/log/llm-red-team.log"

providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
    max_requests_per_minute: 3000
    timeout: 30s
    
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    max_requests_per_minute: 1000
    timeout: 30s
```

#### Docker Deployment
```dockerfile
# Multi-stage build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o llm-red-team ./src/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/llm-red-team .
COPY --from=builder /app/examples ./examples
COPY --from=builder /app/templates ./templates

EXPOSE 8080 8090 6060

CMD ["./llm-red-team", "server", "--config", "/etc/config/production-config.yaml"]
```

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  llm-red-team-1:
    build: .
    environment:
      - NODE_ID=node-1
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./production-config.yaml:/etc/config/production-config.yaml
      - ./certs:/etc/certs
      - logs-1:/var/log
    ports:
      - "8080:8080"
      - "8090:8090"
      - "6060:6060"
    depends_on:
      - redis-cluster
      
  llm-red-team-2:
    build: .
    environment:
      - NODE_ID=node-2
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./production-config.yaml:/etc/config/production-config.yaml
      - ./certs:/etc/certs
      - logs-2:/var/log
    ports:
      - "8081:8080"
      - "8091:8090"
      - "6061:6060"
    depends_on:
      - redis-cluster
      
  llm-red-team-3:
    build: .
    environment:
      - NODE_ID=node-3
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./production-config.yaml:/etc/config/production-config.yaml
      - ./certs:/etc/certs
      - logs-3:/var/log
    ports:
      - "8082:8080"
      - "8092:8090"
      - "6062:6060"
    depends_on:
      - redis-cluster

volumes:
  logs-1:
  logs-2:
  logs-3:
```

#### Kubernetes Deployment
```yaml
# llm-red-team-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: llm-red-team
spec:
  replicas: 3
  selector:
    matchLabels:
      app: llm-red-team
  template:
    metadata:
      labels:
        app: llm-red-team
    spec:
      containers:
      - name: llm-red-team
        image: llm-red-team:v0.2.0
        ports:
        - containerPort: 8080
        - containerPort: 8090
        - containerPort: 6060
        env:
        - name: NODE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: jwt-secret
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai
        volumeMounts:
        - name: config
          mountPath: /etc/config
        - name: certs
          mountPath: /etc/certs
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "8Gi"
            cpu: "4000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: llm-red-team-config
      - name: certs
        secret:
          secretName: tls-certs
---
apiVersion: v1
kind: Service
metadata:
  name: llm-red-team-service
spec:
  selector:
    app: llm-red-team
  ports:
  - name: api
    port: 8080
    targetPort: 8080
  - name: monitoring
    port: 8090
    targetPort: 8090
  - name: profiling
    port: 6060
    targetPort: 6060
  type: ClusterIP
```

### 3. Load Balancer Configuration

#### HAProxy Configuration
```
# haproxy.cfg
global
    daemon
    maxconn 4096
    log stdout local0

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog

frontend llm_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/llm-red-team.pem
    redirect scheme https if !{ ssl_fc }
    default_backend llm_backend

backend llm_backend
    balance roundrobin
    option httpchk GET /health
    server node1 llm-red-team-1:8080 check
    server node2 llm-red-team-2:8080 check
    server node3 llm-red-team-3:8080 check

frontend monitoring_frontend
    bind *:8090
    default_backend monitoring_backend

backend monitoring_backend
    balance roundrobin
    server node1 llm-red-team-1:8090 check
    server node2 llm-red-team-2:8090 check
    server node3 llm-red-team-3:8090 check
```

### 4. Monitoring Setup

#### Prometheus Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: 'llm-red-team'
  static_configs:
  - targets:
    - 'llm-red-team-1:8090'
    - 'llm-red-team-2:8090'
    - 'llm-red-team-3:8090'
  metrics_path: '/api/v1/metrics'
  scrape_interval: 5s

- job_name: 'redis-cluster'
  static_configs:
  - targets:
    - 'redis-1:7000'
    - 'redis-2:7001'
    - 'redis-3:7002'

- job_name: 'profiling'
  static_configs:
  - targets:
    - 'llm-red-team-1:6060'
    - 'llm-red-team-2:6060'
    - 'llm-red-team-3:6060'
  metrics_path: '/debug/pprof/goroutine'
```

#### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "LLM Red Team v0.2.0 Monitoring",
    "panels": [
      {
        "title": "Attack Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(attacks_total[5m])",
            "legendFormat": "Attacks/sec"
          }
        ]
      },
      {
        "title": "Cluster Health",
        "type": "stat",
        "targets": [
          {
            "expr": "cluster_active_nodes",
            "legendFormat": "Active Nodes"
          }
        ]
      },
      {
        "title": "Redis Operations",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(redis_operations_total[5m])",
            "legendFormat": "Redis Ops/sec"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "process_resident_memory_bytes",
            "legendFormat": "Memory {{instance}}"
          }
        ]
      }
    ]
  }
}
```

## Operations

### Scaling Operations

#### Horizontal Scaling
```bash
# Add new node to cluster
docker-compose -f docker-compose.prod.yml up -d --scale llm-red-team=4

# Kubernetes scaling
kubectl scale deployment llm-red-team --replicas=5
```

#### Vertical Scaling
```bash
# Update resource limits
kubectl patch deployment llm-red-team -p '{"spec":{"template":{"spec":{"containers":[{"name":"llm-red-team","resources":{"limits":{"memory":"16Gi","cpu":"8000m"}}}]}}}}'
```

### Maintenance Operations

#### Rolling Updates
```bash
# Docker Compose
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --no-deps --build

# Kubernetes
kubectl set image deployment/llm-red-team llm-red-team=llm-red-team:v0.2.1 --record
kubectl rollout status deployment/llm-red-team
```

#### Backup Operations
```bash
# Redis cluster backup
redis-cli --cluster backup redis-1:7000 --cluster-backup-dir /backup

# Application state backup
kubectl exec -it llm-red-team-pod -- tar czf /backup/app-state.tar.gz /var/lib/llm-red-team
```

### Performance Tuning

#### Redis Optimization
```
# redis.conf optimizations
maxmemory 8gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error no
rdbcompression yes
```

#### Application Tuning
```yaml
# Go runtime optimizations
environment:
  - GOMAXPROCS=8
  - GOGC=100
  - GOMEMLIMIT=8GiB
```

## Security Considerations

### Network Security
- Use private networks for inter-node communication
- Implement TLS for all external communications
- Configure firewalls to restrict access
- Use VPN for administrative access

### Authentication & Authorization
- Enable JWT-based authentication
- Implement RBAC for multi-user access
- Use API keys for service-to-service communication
- Rotate secrets regularly

### Data Protection
- Encrypt data at rest and in transit
- Implement audit logging
- Use secrets management (HashiCorp Vault, K8s Secrets)
- Regular security scanning

## Troubleshooting

### Common Issues

#### Redis Cluster Split-Brain
```bash
# Check cluster status
redis-cli --cluster check redis-1:7000

# Fix cluster
redis-cli --cluster fix redis-1:7000
```

#### High Memory Usage
```bash
# Check memory usage
curl http://localhost:6060/debug/pprof/heap

# Force garbage collection
curl -X POST http://localhost:8090/api/v1/gc
```

#### Performance Degradation
```bash
# Check profiling data
go tool pprof http://localhost:6060/debug/pprof/profile

# Review metrics
curl http://localhost:8090/api/v1/metrics | grep performance
```

### Monitoring Alerts

#### Critical Alerts
- Node down (>30 seconds)
- Redis cluster split-brain
- Memory usage >90%
- Attack success rate <50%

#### Warning Alerts
- High CPU usage (>80%)
- Redis memory usage >80%
- API latency >5 seconds
- Low cache hit ratio (<70%)

## Performance Benchmarks

### Expected Performance (3-node cluster)
- **Concurrent Attacks**: 100-500 simultaneous
- **Throughput**: 50-200 attacks/second
- **Latency**: <2 seconds average response time
- **Availability**: 99.9% uptime
- **Cache Hit Ratio**: >80%

### Scaling Characteristics
- **Linear scaling** up to 10 nodes
- **Memory usage**: ~2GB per 100 concurrent attacks
- **CPU usage**: ~4 cores per 100 concurrent attacks
- **Network**: ~100Mbps per 100 concurrent attacks

This production deployment guide ensures reliable, scalable operation of LLM Red Team v0.2.0 in enterprise environments.