# Performance Optimization Guide - v0.2.0

This guide covers performance optimization techniques for LLMrecon v0.2.0 to achieve maximum throughput and efficiency.

## Overview

v0.2.0 introduces comprehensive performance optimization capabilities:
- **Automated profiling** and hotspot detection
- **Memory optimization** with object pooling
- **Connection pooling** for HTTP clients
- **Distributed caching** with Redis clustering
- **Auto-scaling** based on load metrics
- **Real-time monitoring** and alerting

## Performance Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Performance Optimization Stack               │
├─────────────────────────────────────────────────────────────┤
│ Application Layer                                           │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│ │  Profiler   │ │  Optimizer  │ │  Monitor    │            │
│ │  Engine     │ │  Engine     │ │  Dashboard  │            │
│ └─────────────┘ └─────────────┘ └─────────────┘            │
├─────────────────────────────────────────────────────────────┤
│ Concurrency Layer                                          │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│ │ Worker      │ │ Load        │ │ Rate        │            │
│ │ Pools       │ │ Balancer    │ │ Limiter     │            │
│ └─────────────┘ └─────────────┘ └─────────────┘            │
├─────────────────────────────────────────────────────────────┤
│ Caching Layer                                               │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│ │ Redis       │ │ Memory      │ │ Connection  │            │
│ │ Cluster     │ │ Pools       │ │ Pools       │            │
│ └─────────────┘ └─────────────┘ └─────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

## Performance Profiling

### Automated Profiling Setup

```yaml
# config.yaml
profiler:
  enabled: true
  
  # Profiling intervals
  profiling_interval: 30s
  metrics_interval: 10s
  optimization_interval: 5m
  
  # Enable specific profilers
  cpu_profiling: true
  memory_profiling: true
  goroutine_profiling: true
  block_profiling: true
  mutex_profiling: true
  trace_profiling: false  # Expensive, use sparingly
  
  # Storage settings
  profiles_dir: "./profiles"
  max_profile_files: 100
  profile_retention: 24h
  
  # Server settings
  server_enabled: true
  server_host: "0.0.0.0"
  server_port: 6060
  
  # Analysis settings
  analysis_enabled: true
  analysis_depth: 10
  hotspot_threshold: 5.0
  memory_leak_threshold: 100MB
  
  # Auto-optimization
  auto_optimization: false  # Enable with caution
  optimization_strategies: ["gc_tuning", "pool_optimization"]
  gc_tuning_enabled: true
  pool_optimization_enabled: true
  
  # Alerting
  alerts_enabled: true
  performance_thresholds:
    max_cpu_usage: 80.0
    max_memory_usage: 1GB
    max_goroutines: 10000
    max_latency: 1s
    max_error_rate: 5.0
    min_throughput: 100.0
```

### Manual Profiling Commands

```bash
# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Collect memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Collect goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Collect block profile
go tool pprof http://localhost:6060/debug/pprof/block

# Collect mutex profile
go tool pprof http://localhost:6060/debug/pprof/mutex

# View profiling dashboard
open http://localhost:6060/debug/pprof/
```

### Performance Analysis

```bash
# Generate performance report
./llmrecon report generate \
  --type performance \
  --period 1h \
  --output performance-report.json

# Analyze hotspots
curl http://localhost:8090/api/v1/hotspots | jq '.[0:5]'

# View performance issues
curl http://localhost:8090/api/v1/issues | jq '.[] | select(.severity == "high")'

# Check optimization recommendations
curl http://localhost:8090/api/v1/recommendations
```

## Memory Optimization

### Object Pooling Configuration

```yaml
memory_optimization:
  enabled: true
  
  # Object pooling settings
  byte_slice_pool:
    enabled: true
    sizes: [1024, 8192, 65536]  # Pool different sizes
    max_objects: 1000
    
  string_builder_pool:
    enabled: true
    initial_capacity: 4096
    max_objects: 500
    
  map_pool:
    enabled: true
    types: ["string_interface", "string_string"]
    max_objects: 200
    
  # Garbage collection tuning
  gc_tuning:
    enabled: true
    target_percentage: 100  # GOGC setting
    memory_limit: 8GB       # GOMEMLIMIT setting
    
  # Memory pressure monitoring
  pressure_monitoring:
    enabled: true
    check_interval: 30s
    warning_threshold: 80   # % of memory limit
    critical_threshold: 95  # % of memory limit
```

### Memory Usage Monitoring

```bash
# Check current memory usage
curl http://localhost:8090/api/v1/metrics | jq '.memory'

# Memory allocation stats
curl http://localhost:6060/debug/pprof/heap?debug=1

# Force garbage collection
curl -X POST http://localhost:8090/api/v1/gc

# Memory leak detection
./llmrecon analyze memory-leaks --duration 10m
```

## Connection Pooling

### HTTP Connection Pool Configuration

```yaml
connection_pools:
  enabled: true
  
  # Per-provider pool settings
  providers:
    openai:
      max_connections: 100
      max_idle_connections: 20
      idle_timeout: 90s
      connection_timeout: 30s
      response_timeout: 60s
      health_check_interval: 30s
      
    anthropic:
      max_connections: 50
      max_idle_connections: 10
      idle_timeout: 90s
      connection_timeout: 30s
      response_timeout: 60s
      health_check_interval: 30s
  
  # Global pool settings
  global_limits:
    max_total_connections: 500
    max_connections_per_host: 50
    keep_alive_timeout: 30s
    
  # Health checking
  health_checks:
    enabled: true
    interval: 30s
    timeout: 5s
    failure_threshold: 3
    recovery_timeout: 60s
```

### Connection Pool Monitoring

```bash
# Check pool statistics
curl http://localhost:8090/api/v1/connection-pools

# Monitor connection health
curl http://localhost:8090/api/v1/connection-pools/health

# Pool utilization metrics
curl http://localhost:8090/api/v1/metrics | jq '.connection_pools'
```

## Distributed Caching

### Redis Cluster Cache Configuration

```yaml
caching:
  redis_cluster:
    enabled: true
    
    # Cluster configuration
    nodes:
      - "redis-1:7000"
      - "redis-2:7001"
      - "redis-3:7002"
    password: "${REDIS_PASSWORD}"
    max_redirects: 8
    read_timeout: 3s
    write_timeout: 3s
    
    # Partitioning strategy
    partition_strategy: "consistent"
    partition_count: 256
    replication_factor: 3
    
    # Cache settings
    default_ttl: 1h
    max_value_size: 1MB
    compression_enabled: true
    compression_threshold: 1KB
    
    # Cache warming
    warming_enabled: true
    warming_interval: 10m
    warming_concurrency: 5
    warming_batch_size: 100
    
    # Invalidation
    invalidation_strategy: "ttl"
    tags_enabled: true
    max_tags: 10
    
    # Performance optimization
    pipeline_size: 100
    pool_size: 20
    min_idle_connections: 5
    read_preference: "prefer_replica"
```

### Cache Performance Tuning

```bash
# Cache hit ratio monitoring
curl http://localhost:8090/api/v1/cache/metrics | jq '.hit_ratio'

# Cache warming status
curl http://localhost:8090/api/v1/cache/warming/status

# Invalidate cache by tags
curl -X POST http://localhost:8090/api/v1/cache/invalidate \
  -H "Content-Type: application/json" \
  -d '{"tags": ["attack_results", "provider_responses"]}'

# Cache size and usage
redis-cli --cluster info redis-1:7000
```

## Concurrency Optimization

### Worker Pool Configuration

```yaml
concurrency:
  enabled: true
  
  # Worker pool settings
  default_pool_size: 16      # 2x CPU cores
  max_workers: 100
  min_workers: 4
  worker_idle_timeout: 5m
  
  # Task scheduling
  scheduling_algorithm: "adaptive"
  priority_levels: 5
  task_timeout: 30s
  max_queue_size: 1000
  
  # Load balancing
  balancing_strategy: "adaptive"
  health_check_interval: 10s
  circuit_breaker:
    failure_threshold: 5
    recovery_timeout: 30s
    half_open_requests: 3
  
  # Adaptive scaling
  adaptive_scaling:
    enabled: true
    scaling_interval: 30s
    cpu_threshold: 80.0
    memory_threshold: 85.0
  
  # Coordination (distributed mode)
  coordination:
    enabled: true
    mode: "hybrid"
    heartbeat_interval: 5s
```

### Concurrency Monitoring

```bash
# Worker pool status
curl http://localhost:8090/api/v1/workers/status

# Task queue depth
curl http://localhost:8090/api/v1/metrics | jq '.tasks_queued'

# Scaling events
curl http://localhost:8090/api/v1/scaling/events

# Concurrency metrics
curl http://localhost:8090/api/v1/metrics | jq '.concurrency'
```

## Rate Limiting Optimization

### Distributed Rate Limiter Configuration

```yaml
rate_limiting:
  enabled: true
  
  # Redis configuration
  redis_addr: "redis-cluster:6379"
  redis_password: "${REDIS_PASSWORD}"
  redis_db: 0
  
  # Rate limiting settings
  key_prefix: "ratelimit"
  default_limit: 1000
  default_window: 60s
  default_burst: 10
  
  # Algorithm settings
  algorithm: "token_bucket"  # token_bucket, sliding_window, fixed_window, leaky_bucket
  sliding_window_parts: 10
  
  # Cleanup and maintenance
  cleanup_interval: 5m
  key_expiration: 1h
  
  # Performance settings
  enable_pipelining: true
  max_retries: 3
  retry_delay: 100ms
  
  # Per-provider limits
  provider_limits:
    openai:
      limit: 3000
      window: 60s
      burst: 50
      
    anthropic:
      limit: 1000
      window: 60s
      burst: 20
```

### Rate Limiting Monitoring

```bash
# Rate limit status
curl http://localhost:8090/api/v1/rate-limit/status

# Rate limit metrics
curl http://localhost:8090/api/v1/metrics | jq '.rate_limiting'

# Check specific key status
curl "http://localhost:8090/api/v1/rate-limit/status?key=provider:openai"
```

## Load Balancing

### Advanced Load Balancer Configuration

```yaml
load_balancing:
  enabled: true
  
  # Strategy configuration
  default_strategy: "adaptive"
  
  # Health monitoring
  health_checks:
    enabled: true
    interval: 10s
    timeout: 5s
    failure_threshold: 3
    
  # Circuit breaker
  circuit_breaker:
    failure_threshold: 5
    recovery_timeout: 30s
    half_open_requests: 3
    
  # Auto-scaling
  auto_scaling:
    enabled: true
    min_targets: 1
    max_targets: 10
    scale_up_threshold: 80.0
    scale_down_threshold: 30.0
    cooldown_period: 5m
    
  # Load prediction
  load_prediction:
    enabled: true
    window_size: 100
    prediction_horizon: 5m
    algorithms: ["linear", "exponential", "seasonal"]
```

### Load Balancing Monitoring

```bash
# Load balancer status
curl http://localhost:8090/api/v1/load-balancer/status

# Target health
curl http://localhost:8090/api/v1/load-balancer/targets

# Load distribution metrics
curl http://localhost:8090/api/v1/metrics | jq '.load_balancing'
```

## Performance Benchmarking

### Built-in Benchmarks

```bash
# Run memory benchmark
./llmrecon benchmark memory \
  --duration 5m \
  --concurrent-workers 50 \
  --output memory-benchmark.json

# Run execution benchmark
./llmrecon benchmark execution \
  --attacks 1000 \
  --concurrent 50 \
  --providers openai,anthropic \
  --output execution-benchmark.json

# Run caching benchmark
./llmrecon benchmark cache \
  --operations 10000 \
  --key-size 100 \
  --value-size 1KB \
  --output cache-benchmark.json

# Run distributed execution benchmark
./llm-red-team benchmark distributed \
  --nodes 3 \
  --attacks-per-node 500 \
  --coordination-overhead \
  --output distributed-benchmark.json
```

### Custom Benchmarks

```bash
# Create custom benchmark
./llm-red-team benchmark create \
  --name "custom-attack-benchmark" \
  --template "examples/benchmarks/attack-template.yaml" \
  --config "benchmarks/custom-config.yaml"

# Run custom benchmark
./llm-red-team benchmark run custom-attack-benchmark \
  --iterations 1000 \
  --concurrent 100 \
  --output custom-results.json
```

## Performance Tuning Strategies

### CPU Optimization

1. **Goroutine Management**:
   ```yaml
   # Optimal GOMAXPROCS setting
   environment:
     GOMAXPROCS: "16"  # Number of CPU cores
   ```

2. **CPU Profiling Analysis**:
   ```bash
   # Identify CPU hotspots
   go tool pprof -top http://localhost:6060/debug/pprof/profile
   
   # Interactive CPU analysis
   go tool pprof http://localhost:6060/debug/pprof/profile
   (pprof) top 10
   (pprof) list hotspot_function
   ```

3. **CPU-Intensive Task Optimization**:
   - Use worker pools for CPU-bound tasks
   - Implement efficient algorithms
   - Minimize context switching

### Memory Optimization

1. **Memory Pool Tuning**:
   ```yaml
   memory_pools:
     # Size pools based on actual usage patterns
     byte_slice_sizes: [512, 4096, 32768, 262144]
     
     # Monitor pool hit rates
     monitoring:
       enabled: true
       hit_rate_threshold: 80.0
   ```

2. **Garbage Collection Tuning**:
   ```bash
   # Aggressive GC for low latency
   export GOGC=50
   
   # Conservative GC for high throughput
   export GOGC=200
   
   # Memory limit setting
   export GOMEMLIMIT=8GiB
   ```

3. **Memory Leak Detection**:
   ```bash
   # Monitor memory growth
   watch -n 5 'curl -s http://localhost:8090/api/v1/metrics | jq .memory.heap_alloc'
   
   # Analyze memory allocation
   go tool pprof -alloc_space http://localhost:6060/debug/pprof/heap
   ```

### Network Optimization

1. **Connection Pool Tuning**:
   ```yaml
   # Optimize for high throughput
   connection_pools:
     max_connections: 200
     max_idle_connections: 50
     idle_timeout: 30s
     
   # Optimize for low latency
   connection_pools:
     max_connections: 100
     max_idle_connections: 20
     idle_timeout: 10s
   ```

2. **TCP Settings**:
   ```bash
   # Linux kernel tuning
   echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
   echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
   echo 'net.ipv4.tcp_rmem = 4096 65536 16777216' >> /etc/sysctl.conf
   echo 'net.ipv4.tcp_wmem = 4096 65536 16777216' >> /etc/sysctl.conf
   sysctl -p
   ```

### Database/Cache Optimization

1. **Redis Cluster Tuning**:
   ```
   # redis.conf optimizations
   maxmemory-policy allkeys-lru
   save ""  # Disable RDB for performance
   appendonly yes
   appendfsync everysec
   tcp-keepalive 60
   timeout 300
   ```

2. **Cache Strategy Optimization**:
   ```yaml
   caching:
     # Use appropriate TTLs
     attack_results_ttl: 1h
     provider_responses_ttl: 30m
     
     # Implement cache warming
     warming_strategies:
       - "lru"
       - "frequency"
       - "predictive"
   ```

## Performance Monitoring and Alerting

### Key Performance Indicators (KPIs)

1. **Throughput Metrics**:
   - Attacks per second
   - Requests per second per provider
   - Cache operations per second

2. **Latency Metrics**:
   - Average response time
   - P95/P99 latency
   - Queue wait time

3. **Resource Utilization**:
   - CPU usage
   - Memory usage
   - Network bandwidth
   - Goroutine count

4. **Error Rates**:
   - Attack failure rate
   - Provider error rate
   - Cache miss rate

### Alerting Rules

```yaml
# Prometheus alerting rules
groups:
- name: llm-red-team-performance
  rules:
  - alert: HighCPUUsage
    expr: cpu_usage_percent > 80
    for: 5m
    annotations:
      summary: "High CPU usage detected"
      
  - alert: HighMemoryUsage
    expr: memory_usage_bytes / memory_limit_bytes > 0.85
    for: 5m
    annotations:
      summary: "High memory usage detected"
      
  - alert: LowThroughput
    expr: rate(attacks_total[5m]) < 10
    for: 10m
    annotations:
      summary: "Low attack throughput detected"
      
  - alert: HighLatency
    expr: histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m])) > 5
    for: 5m
    annotations:
      summary: "High request latency detected"
```

## Troubleshooting Performance Issues

### Common Performance Problems

1. **High CPU Usage**:
   ```bash
   # Identify CPU hotspots
   go tool pprof -top http://localhost:6060/debug/pprof/profile
   
   # Check goroutine blocking
   go tool pprof http://localhost:6060/debug/pprof/block
   ```

2. **Memory Leaks**:
   ```bash
   # Monitor memory growth
   curl http://localhost:8090/api/v1/metrics | jq '.memory.heap_alloc'
   
   # Analyze allocation patterns
   go tool pprof -alloc_space http://localhost:6060/debug/pprof/heap
   ```

3. **Slow Response Times**:
   ```bash
   # Check request latency
   curl http://localhost:8090/api/v1/metrics | jq '.latency'
   
   # Analyze request traces
   curl http://localhost:6060/debug/pprof/trace?seconds=10
   ```

4. **Low Throughput**:
   ```bash
   # Check worker utilization
   curl http://localhost:8090/api/v1/workers/status
   
   # Monitor queue depth
   curl http://localhost:8090/api/v1/metrics | jq '.queue_depth'
   ```

### Performance Optimization Checklist

- [ ] Enable performance profiling
- [ ] Configure appropriate worker pool sizes
- [ ] Set up connection pooling
- [ ] Enable Redis cluster caching
- [ ] Configure memory optimization
- [ ] Set up rate limiting
- [ ] Enable load balancing
- [ ] Configure monitoring and alerting
- [ ] Run performance benchmarks
- [ ] Optimize based on profiling data
- [ ] Monitor key performance metrics
- [ ] Set up automated scaling

This performance optimization guide provides comprehensive strategies for maximizing the performance of LLM Red Team v0.2.0 in production environments.