# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

### Building the main application
```bash
# Build the main CLI tool
go build -o llmrecon ./src/main.go

# Build specific tools
go build -o compliance-report ./cmd/compliance-report
go build -o template_security ./cmd/template_security_standalone/main.go
go build -o config-manager ./cmd/config-manager
go build -o execution-benchmark ./cmd/execution-benchmark
go build -o cache-benchmark ./cmd/cache-benchmark
go build -o owasp-mock-test ./cmd/owasp-mock-test

# Build individual components
./scripts/build_component.sh template_security
./scripts/build_component.sh audit_logger
./scripts/build_component.sh memory_optimizer
./scripts/build_component.sh monitoring
./scripts/build_component.sh performance_profiler
./scripts/build_component.sh distributed_coordinator
```

### Running tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/template/...
go test ./src/security/...
go test ./src/bundle/...

# Run benchmarks
./scripts/run_memory_benchmark.sh
./scripts/run_execution_benchmark.sh
./scripts/benchmark_caching.sh
./scripts/benchmark_executors.sh
./scripts/benchmark_redis_cluster.sh
./scripts/benchmark_distributed_execution.sh
```

### Template development and validation
```bash
# Verify template compliance
./scripts/verify-compliance.sh

# Optimize templates
./scripts/optimize_templates.sh

# Run template security checks
go run ./cmd/template_security_standalone/main.go
```

## Architecture Overview

This is an enterprise-grade LLM security testing tool implementing OWASP LLM Top 10 and ISO/IEC 42001 compliance frameworks.

### Core Architecture Patterns

1. **Layered Architecture**:
   - CLI Layer (`src/cmd/`) - Cobra-based command interface
   - API Layer (`src/api/`) - RESTful API with Gorilla Mux
   - Business Logic (`src/`) - Core functionality organized by domain
   - Repository Pattern - Abstraction for storage backends (GitHub, GitLab, S3, local)

2. **Plugin System**:
   - Provider plugins for LLM APIs (OpenAI, Anthropic, etc.)
   - Dynamic loading with version compatibility checking
   - Located in `src/provider/` with factory pattern for instantiation

3. **Template Engine**:
   - YAML-based vulnerability test templates (similar to Nuclei)
   - Template validation, caching, and execution pipeline
   - Templates organized by OWASP categories in `examples/templates/owasp-llm/`

4. **Security Framework**:
   - RBAC with multi-factor authentication support
   - Audit trail management with structured logging
   - Secure communication with TLS and certificate management
   - Prompt injection protection and content filtering

### Key Components and Relationships

1. **Template Management System** (`src/template/`):
   - Loads and validates YAML templates
   - Manages template execution with rate limiting
   - Caches compiled templates for performance
   - Supports inheritance and modular composition

2. **Provider Framework** (`src/provider/`):
   - Interface-based design for LLM provider integration
   - Middleware stack: rate limiting, retries, circuit breaker, logging
   - Configuration management with encryption for sensitive data

3. **Update System** (`src/update/`):
   - Self-updating capability for binary and templates
   - Version management with semantic versioning
   - Signature verification for secure updates
   - Offline bundle support for air-gapped environments

4. **Reporting System** (`src/reporting/`):
   - Multiple output formats via factory pattern
   - Compliance-focused reporting for OWASP and ISO standards
   - Integration with vulnerability management systems

5. **Bundle System** (`src/bundle/`):
   - Offline distribution format for templates and modules
   - Conflict resolution for template updates
   - Import/export with validation and rollback support

### Important Design Decisions

1. **Interface-Heavy Design**: Most components define interfaces first, implementations second. This enables easy testing and extensibility.

2. **Factory Pattern Usage**: Providers, reports, and many other components use factories for instantiation, supporting runtime configuration.

3. **Middleware Architecture**: API and provider calls go through configurable middleware stacks for cross-cutting concerns.

4. **Template Inheritance**: Templates can inherit from base templates, promoting reuse and consistency.

5. **Multi-Repository Support**: Can sync templates from multiple sources (GitHub for production, GitLab for development).

### Current Development Focus

The codebase shows active development on:
- OWASP LLM Top 10 compliance implementation
- Production-scale infrastructure (v0.2.0)
- Distributed execution and coordination
- Advanced caching and performance optimization
- Real-time monitoring and profiling
- Memory optimization for large-scale operations
- Enhanced template security verification
- Offline bundle functionality
- Access control and authentication improvements

### v0.2.0 Production Scale Infrastructure

Version 0.2.0 introduces enterprise-grade infrastructure for scaling from ~10 to 100+ concurrent attacks:

1. **HTTP Connection Pooling** (`src/provider/core/connection_pool.go`):
   - Per-provider connection pools with health checks
   - Connection reuse and lifecycle management
   - Automatic failover and recovery

2. **Redis-Backed Job Queue** (`src/queue/`):
   - Persistent job queue using Redis sorted sets
   - Priority-based job scheduling with retry logic
   - Worker management with auto-scaling

3. **Memory Optimization** (`src/performance/memory_pool.go`):
   - Object pooling for high-frequency allocations
   - Automatic cleanup and GC optimization
   - Memory pressure monitoring

4. **Distributed Rate Limiting** (`src/performance/distributed_rate_limiter.go`):
   - Redis-based rate limiting with Lua scripts
   - Multiple algorithms: token bucket, sliding window, fixed window, leaky bucket
   - Atomic operations and distributed coordination

5. **Real-Time Monitoring Dashboard** (`src/performance/monitoring_dashboard.go`):
   - WebSocket-based real-time updates
   - REST API endpoints for metrics
   - Multi-client support with authentication

6. **Advanced Concurrency Engine** (`src/performance/concurrency_engine.go`):
   - Worker pools with adaptive scaling
   - Task scheduling with multiple algorithms
   - Pipeline execution patterns

7. **Load Balancing & Auto-Scaling** (`src/performance/load_balancer.go`):
   - Multiple load balancing strategies
   - Health monitoring and circuit breakers
   - Predictive auto-scaling capabilities

8. **Distributed Execution Coordinator** (`src/performance/distributed_coordinator.go`):
   - Multi-node task distribution
   - Leader election and consensus management
   - Task partitioning and replication

9. **Advanced Redis Cluster Cache** (`src/performance/redis_cluster_cache.go`):
   - Redis cluster support with partitioning
   - Cache warming and invalidation strategies
   - Tag-based cache management

10. **Performance Profiling System** (`src/performance/profiler.go`):
    - Comprehensive CPU, memory, goroutine, block, and mutex profiling
    - Automated performance analysis and hotspot detection
    - Real-time optimization recommendations

### Dependencies and Infrastructure Requirements

v0.2.0 requires additional dependencies for distributed execution:

```bash
# Core dependencies
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/go-sql-driver/mysql
go get github.com/lib/pq
go get golang.org/x/term

# v0.2.0 Production Scale dependencies
go get github.com/go-redis/redis/v8
go get github.com/gorilla/mux
go get github.com/gorilla/websocket
```

### Infrastructure Requirements for v0.2.0

For production-scale deployment:

1. **Redis Cluster**:
   - Minimum 3-node Redis cluster for distributed operations
   - Recommended: 6 nodes (3 masters + 3 replicas)
   - Memory: 8GB+ per node for caching

2. **Application Nodes**:
   - CPU: 8+ cores per node
   - Memory: 16GB+ per node  
   - Network: Low latency between nodes
   - Recommended: 3+ nodes for high availability

3. **Monitoring Infrastructure**:
   - Prometheus/Grafana for metrics (optional)
   - Log aggregation system (ELK/Loki)
   - Alert manager for notifications

### Configuration Examples

#### Redis Cluster Configuration
```yaml
redis:
  cluster:
    nodes:
      - "redis-1:7000"
      - "redis-2:7000" 
      - "redis-3:7000"
    password: "${REDIS_PASSWORD}"
    max_redirects: 8
    read_timeout: 3s
    write_timeout: 3s
```

#### Performance Profiler Configuration
```yaml
profiler:
  enabled: true
  server:
    host: "0.0.0.0"
    port: 6060
  profiling:
    cpu_enabled: true
    memory_enabled: true
    interval: 30s
  optimization:
    auto_enabled: false
    strategies: ["gc_tuning", "pool_optimization"]
```

#### Distributed Coordinator Configuration
```yaml
coordinator:
  node_id: "node-1"
  cluster_name: "llm-cluster"
  redis_addr: "redis-cluster:6379"
  heartbeat_interval: 5s
  task_partitioning: true
  replication_factor: 3
```