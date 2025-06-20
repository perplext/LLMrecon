# Distributed Architecture Guide - v0.2.0

This guide explains the distributed architecture of LLM Red Team v0.2.0 and how to design, deploy, and manage distributed attack campaigns at scale.

## Overview

v0.2.0 introduces a sophisticated distributed architecture that enables:
- **Horizontal scaling** across multiple nodes
- **Distributed coordination** with consensus mechanisms  
- **Load balancing** with intelligent request routing
- **Fault tolerance** with automatic failover
- **State synchronization** across the cluster
- **100+ concurrent attacks** coordinated across nodes

## Architecture Components

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Distributed LLM Red Team                     │
├─────────────────────────────────────────────────────────────────┤
│                      Load Balancer                              │
│              ┌─────────────────────────────┐                     │
│              │    HAProxy / NGINX / ALB    │                     │
│              └─────────────────────────────┘                     │
├─────────────────────────────────────────────────────────────────┤
│        Application Cluster (3+ Nodes)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Node 1    │  │   Node 2    │  │   Node 3    │              │
│  │ (Leader)    │  │ (Follower)  │  │ (Follower)  │              │
│  │             │  │             │  │             │              │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │              │
│  │ │Attack   │ │  │ │Attack   │ │  │ │Attack   │ │              │
│  │ │Engine   │ │  │ │Engine   │ │  │ │Engine   │ │              │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │              │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │              │
│  │ │Job      │ │  │ │Job      │ │  │ │Job      │ │              │
│  │ │Workers  │ │  │ │Workers  │ │  │ │Workers  │ │              │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │              │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │              │
│  │ │Monitor  │ │  │ │Monitor  │ │  │ │Monitor  │ │              │
│  │ │Dashboard│ │  │ │Dashboard│ │  │ │Dashboard│ │              │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
├─────────────────────────────────────────────────────────────────┤
│                    Redis Cluster                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ Master:1    │  │ Master:2    │  │ Master:3    │              │
│  │ Slave:4     │  │ Slave:5     │  │ Slave:6     │              │
│  │             │  │             │  │             │              │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │              │
│  │ │Job Queue│ │  │ │Cache    │ │  │ │State    │ │              │
│  │ │Rate Lmt │ │  │ │Results  │ │  │ │Coord    │ │              │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. Distributed Execution Coordinator
**File**: `src/performance/distributed_coordinator.go`

**Responsibilities**:
- Cluster membership management
- Leader election and consensus
- Task distribution and partitioning
- Node health monitoring
- Failover coordination

**Key Features**:
```go
type DistributedExecutionCoordinator struct {
    nodeManager     *NodeManager      // Manages cluster nodes
    taskOrchestrator *TaskOrchestrator // Distributes tasks
    consensus       *ConsensusManager  // Handles consensus
    election        *LeaderElection   // Manages leader election
    partition       *PartitionManager // Partitions tasks
    replication     *ReplicationManager // Replicates data
}
```

#### 2. Job Queue System
**Files**: `src/queue/redis_queue.go`, `src/queue/attack_job_handler.go`

**Responsibilities**:
- Persistent job storage using Redis
- Priority-based job scheduling
- Worker management and scaling
- Retry logic and dead letter queues

**Key Features**:
```go
type RedisJobQueue struct {
    redis       *redis.Client
    workers     []*JobWorker
    scheduler   *JobScheduler
    handlers    map[string]JobHandler
}
```

#### 3. Load Balancer
**File**: `src/performance/load_balancer.go`

**Responsibilities**:
- Request distribution across nodes
- Health monitoring and circuit breaking
- Auto-scaling based on load metrics
- Multiple balancing strategies

**Key Features**:
```go
type AdvancedLoadBalancer struct {
    targets     map[string]*LoadBalanceTarget
    strategies  map[string]BalancingStrategy
    health      *HealthMonitor
    scaler      *AutoScaler
    predictor   *LoadPredictor
    circuit     *CircuitBreaker
}
```

#### 4. Distributed Cache
**File**: `src/performance/redis_cluster_cache.go`

**Responsibilities**:
- Distributed caching with Redis clustering
- Cache partitioning and replication
- Cache warming and invalidation
- Tag-based cache management

## Distributed Coordination Patterns

### Leader Election

The cluster uses a Raft-based consensus algorithm for leader election:

```yaml
# Leader election configuration
leader_election:
  algorithm: "raft"
  election_timeout: 10s
  heartbeat_interval: 2s
  term_timeout: 20s
  max_terms: 1000
```

**Election Process**:
1. **Follower State**: Nodes start as followers
2. **Candidate State**: Timeout triggers candidate election
3. **Leader State**: Majority vote establishes leader
4. **Heartbeat**: Leader sends periodic heartbeats
5. **Re-election**: Timeout triggers new election

### Consensus Mechanisms

**Raft Consensus Implementation**:
```go
type ConsensusManager struct {
    config     ConsensusConfig
    proposals  map[string]*Proposal
    votes      map[string]*VoteRecord
    log        *ConsensusLog
    state      ConsensusState
    term       int64
    votedFor   string
}
```

**Proposal Types**:
- Leader election
- Configuration changes
- Task assignments
- Node membership
- Resource rebalancing

### Task Partitioning

**Partitioning Strategies**:

1. **Hash-Based Partitioning**:
   ```go
   func (p *PartitionManager) HashPartition(key string) int {
       hash := crc32.ChecksumIEEE([]byte(key))
       return int(hash) % p.partitionCount
   }
   ```

2. **Range-Based Partitioning**:
   ```go
   func (p *PartitionManager) RangePartition(key string) int {
       // Partition based on key ranges
       return p.getPartitionForRange(key)
   }
   ```

3. **Load-Based Partitioning**:
   ```go
   func (p *PartitionManager) LoadPartition(key string) int {
       // Choose least loaded partition
       return p.getLeastLoadedPartition()
   }
   ```

4. **Consistent Hashing**:
   ```go
   func (p *PartitionManager) ConsistentHashPartition(key string) int {
       return p.hashRing.GetNode(key)
   }
   ```

## Distributed Job Processing

### Job Distribution Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Job Distribution Flow                         │
├─────────────────────────────────────────────────────────────┤
│ 1. Campaign Submission                                      │
│    ┌─────────────┐                                          │
│    │   Client    │ ──► Submit Attack Campaign               │
│    └─────────────┘                                          │
│           │                                                 │
│ 2. Leader Processing                                        │
│    ┌─────────────┐                                          │
│    │   Leader    │ ──► Parse & Partition Campaign           │
│    │    Node     │ ──► Create Job Queue Entries            │
│    └─────────────┘                                          │
│           │                                                 │
│ 3. Job Queue Distribution                                   │
│    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│    │ Redis       │  │ Redis       │  │ Redis       │        │
│    │ Partition 1 │  │ Partition 2 │  │ Partition 3 │        │
│    └─────────────┘  └─────────────┘  └─────────────┘        │
│           │                 │                 │             │
│ 4. Worker Processing                                        │
│    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│    │   Node 1    │  │   Node 2    │  │   Node 3    │        │
│    │  Workers    │  │  Workers    │  │  Workers    │        │
│    └─────────────┘  └─────────────┘  └─────────────┘        │
│           │                 │                 │             │
│ 5. Result Aggregation                                       │
│    ┌─────────────┐                                          │
│    │   Leader    │ ◄── Collect & Aggregate Results          │
│    │    Node     │ ──► Generate Final Report                │
│    └─────────────┘                                          │
└─────────────────────────────────────────────────────────────┘
```

### Job Types and Handlers

```go
// Job types supported by the distributed system
type JobType string

const (
    JobTypeAttack          JobType = "attack"
    JobTypeBatchAttack     JobType = "batch_attack"
    JobTypeTemplateValidation JobType = "template_validation"
    JobTypeProviderTest    JobType = "provider_test"
    JobTypeComplianceCheck JobType = "compliance_check"
)

// Job handler interface
type JobHandler interface {
    Handle(ctx context.Context, job *Job) (*JobResult, error)
    GetSupportedTypes() []JobType
    GetConcurrency() int
}
```

### Distributed Campaign Execution

**Campaign Configuration**:
```yaml
# distributed-campaign.yaml
campaign:
  name: "large-scale-jailbreak-test"
  distributed: true
  
  # Distribution settings
  distribution:
    strategy: "load_balanced"
    max_nodes: 5
    attacks_per_node: 200
    coordination_mode: "async"
    
  # Attack configuration
  attacks:
    - template: "jailbreak-dan-v1"
      count: 500
      concurrent: 50
    - template: "prompt-injection-unicode"
      count: 300
      concurrent: 30
    - template: "hierarchy-override"
      count: 200
      concurrent: 20
      
  # Target configuration
  targets:
    - provider: "openai"
      model: "gpt-4"
      weight: 60
    - provider: "anthropic"
      model: "claude-3"
      weight: 40
      
  # Performance settings
  performance:
    timeout: 30s
    retry_attempts: 3
    rate_limit: 100/minute
    
  # Result collection
  results:
    aggregation_strategy: "merge"
    export_format: "json"
    include_metrics: true
```

**Campaign Execution**:
```bash
# Submit distributed campaign
./llm-red-team campaign submit \
  --config distributed-campaign.yaml \
  --distributed \
  --nodes 3 \
  --monitor \
  --output campaign-results.json

# Monitor campaign progress
./llm-red-team campaign status campaign-12345 --distributed

# Scale campaign mid-execution
./llm-red-team campaign scale campaign-12345 --nodes 5
```

## Node Management

### Node Discovery

**Service Discovery Configuration**:
```yaml
service_discovery:
  enabled: true
  method: "consul"  # consul, etcd, kubernetes, static
  
  consul:
    address: "consul:8500"
    service_name: "llm-red-team"
    health_check:
      interval: 10s
      timeout: 3s
      
  kubernetes:
    namespace: "llm-red-team"
    label_selector: "app=llm-red-team"
    
  static:
    nodes:
      - "node-1:8080"
      - "node-2:8080"
      - "node-3:8080"
```

### Node Health Monitoring

**Health Check Implementation**:
```go
type HealthChecker struct {
    checks map[string]HealthCheck
    config HealthCheckConfig
}

type HealthCheck interface {
    Check(ctx context.Context) HealthStatus
    GetName() string
    GetTimeout() time.Duration
}

// Built-in health checks
var DefaultHealthChecks = []HealthCheck{
    &CPUHealthCheck{threshold: 90.0},
    &MemoryHealthCheck{threshold: 85.0},
    &DiskHealthCheck{threshold: 80.0},
    &RedisHealthCheck{},
    &ProviderHealthCheck{},
}
```

**Health Status Types**:
```go
type HealthStatus string

const (
    HealthStatusHealthy   HealthStatus = "healthy"
    HealthStatusDegraded  HealthStatus = "degraded"
    HealthStatusUnhealthy HealthStatus = "unhealthy"
    HealthStatusUnknown   HealthStatus = "unknown"
)
```

### Auto-Scaling

**Scaling Configuration**:
```yaml
auto_scaling:
  enabled: true
  
  # Scaling metrics
  metrics:
    cpu_threshold: 70.0
    memory_threshold: 80.0
    queue_depth_threshold: 1000
    response_time_threshold: 5s
    
  # Scaling behavior
  scale_up:
    min_nodes: 1
    max_nodes: 10
    step_size: 1
    cooldown_period: 5m
    
  scale_down:
    step_size: 1
    cooldown_period: 10m
    safety_margin: 20.0  # Don't scale down below this margin
    
  # Predictive scaling
  predictive:
    enabled: true
    window_size: 24h
    prediction_horizon: 1h
    algorithms: ["linear", "seasonal"]
```

## Fault Tolerance

### Circuit Breaker Pattern

**Circuit Breaker Configuration**:
```yaml
circuit_breaker:
  enabled: true
  
  # Failure thresholds
  failure_threshold: 10     # Number of failures to open circuit
  success_threshold: 5      # Number of successes to close circuit
  timeout: 60s             # Time to wait before trying again
  
  # Monitoring
  monitoring:
    enabled: true
    window_size: 100        # Number of requests to track
    minimum_requests: 20    # Minimum requests before evaluation
```

**Circuit Breaker States**:
1. **Closed**: Normal operation, requests pass through
2. **Open**: Circuit is open, requests fail fast
3. **Half-Open**: Testing if service has recovered

### Graceful Degradation

**Degradation Strategies**:
```go
type DegradationStrategy interface {
    ShouldDegrade(metrics *SystemMetrics) bool
    Degrade(ctx context.Context) error
    Recover(ctx context.Context) error
}

// Example strategies
var DegradationStrategies = []DegradationStrategy{
    &ReduceConcurrencyStrategy{targetReduction: 50},
    &DisableNonEssentialFeaturesStrategy{},
    &EnableCachingStrategy{aggressiveTTL: true},
    &ReduceProviderTimeoutsStrategy{},
}
```

### Data Replication

**Replication Configuration**:
```yaml
replication:
  enabled: true
  factor: 3              # Number of replicas
  consistency: "eventual" # eventual, strong, session
  
  # Conflict resolution
  conflict_resolution: "last_write_wins"  # lww, vector_clock, manual
  
  # Synchronization
  sync_interval: 10s
  batch_size: 100
  compression: true
```

## Performance at Scale

### Scaling Characteristics

**Linear Scaling Performance**:
- **1 Node**: ~50 concurrent attacks
- **3 Nodes**: ~150 concurrent attacks  
- **5 Nodes**: ~250 concurrent attacks
- **10 Nodes**: ~500 concurrent attacks

**Coordination Overhead**:
- **Leader Election**: 100-500ms
- **Consensus Operations**: 10-50ms
- **Task Distribution**: 5-20ms per task
- **Health Checks**: 1-5ms per node

### Network Optimization

**Network Configuration**:
```yaml
network:
  # Connection pooling
  connection_pool:
    max_connections: 100
    max_idle_connections: 20
    idle_timeout: 90s
    
  # Keep-alive settings
  keep_alive:
    enabled: true
    interval: 30s
    count: 3
    
  # Compression
  compression:
    enabled: true
    algorithms: ["gzip", "deflate"]
    min_size: 1024
    
  # Timeouts
  timeouts:
    connection: 10s
    request: 30s
    response: 60s
```

## Monitoring Distributed Operations

### Distributed Metrics

**Key Metrics to Monitor**:

1. **Cluster Health**:
   - Active nodes
   - Leader election frequency
   - Consensus operation latency
   - Node failure rate

2. **Task Distribution**:
   - Task queue depth per partition
   - Task execution time
   - Task failure rate
   - Load distribution variance

3. **Network Performance**:
   - Inter-node latency
   - Network bandwidth utilization
   - Connection pool utilization
   - Circuit breaker state

### Distributed Tracing

**Tracing Configuration**:
```yaml
tracing:
  enabled: true
  provider: "jaeger"  # jaeger, zipkin, datadog
  
  jaeger:
    endpoint: "http://jaeger:14268/api/traces"
    service_name: "llm-red-team"
    
  # Sampling
  sampling:
    strategy: "probabilistic"
    rate: 0.1  # Sample 10% of requests
    
  # Tags
  tags:
    cluster_name: "${CLUSTER_NAME}"
    node_id: "${NODE_ID}"
    version: "v0.2.0"
```

### Centralized Logging

**Log Aggregation**:
```yaml
logging:
  distributed: true
  
  # Log shipping
  shipping:
    enabled: true
    destination: "elasticsearch"
    endpoint: "http://elasticsearch:9200"
    index_pattern: "llm-red-team-%Y.%m.%d"
    
  # Log correlation
  correlation:
    enabled: true
    trace_id_header: "X-Trace-ID"
    span_id_header: "X-Span-ID"
    
  # Structured logging
  format: "json"
  fields:
    node_id: "${NODE_ID}"
    cluster_name: "${CLUSTER_NAME}"
    component: "${COMPONENT}"
```

## Deployment Patterns

### Blue-Green Deployment

**Strategy**: Deploy new version alongside old version, then switch traffic

```bash
# Deploy green environment
kubectl apply -f deployments/llm-red-team-green.yaml

# Verify green environment
./scripts/verify-deployment.sh green

# Switch traffic to green
kubectl patch service llm-red-team -p '{"spec":{"selector":{"version":"green"}}}'

# Verify traffic switch
./scripts/verify-traffic.sh green

# Clean up blue environment
kubectl delete -f deployments/llm-red-team-blue.yaml
```

### Rolling Deployment

**Strategy**: Gradually replace instances with new version

```bash
# Start rolling update
kubectl set image deployment/llm-red-team llm-red-team=llm-red-team:v0.2.1

# Monitor rollout
kubectl rollout status deployment/llm-red-team

# Rollback if needed
kubectl rollout undo deployment/llm-red-team
```

### Canary Deployment

**Strategy**: Route small percentage of traffic to new version

```yaml
# Canary deployment configuration
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: llm-red-team
spec:
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 300s}
      - setWeight: 50
      - pause: {duration: 300s}
      - setWeight: 100
```

## Security in Distributed Mode

### Inter-Node Authentication

**mTLS Configuration**:
```yaml
security:
  mtls:
    enabled: true
    cert_file: "/etc/certs/node.crt"
    key_file: "/etc/certs/node.key"
    ca_file: "/etc/certs/ca.crt"
    
  # Node authentication
  node_auth:
    method: "certificate"  # certificate, token, oidc
    verify_hostname: true
    allowed_cns: ["llm-red-team-node"]
```

### Network Security

**Network Policies**:
```yaml
# Kubernetes network policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: llm-red-team-netpol
spec:
  podSelector:
    matchLabels:
      app: llm-red-team
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: llm-red-team
    - podSelector:
        matchLabels:
          app: load-balancer
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: redis-cluster
```

This distributed architecture guide provides comprehensive coverage of designing, deploying, and managing LLM Red Team v0.2.0 in distributed environments for maximum scale and reliability.