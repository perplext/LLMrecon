package performance

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancerImpl manages load distribution and auto-scaling
type LoadBalancerImpl struct {
	config     LoadBalancerConfig
	logger     Logger
	algorithms map[string]LoadBalancingAlgorithm
	nodes      map[string]*LoadBalancerNode
	health     *HealthChecker
	scaler     *AutoScaler
	router     *RequestRouter
	predictor  *LoadPredictor
	analyzer   *LoadAnalyzer
	metrics    *LoadBalancerMetrics
	stats      *LoadBalancerStats
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// LoadBalancingAlgorithm interface for different load balancing strategies
type LoadBalancingAlgorithm interface {
	SelectNode(nodes []*LoadBalancerNode, request *Request) (*LoadBalancerNode, error)
	GetName() string
	Configure(config map[string]interface{}) error
}

// LoadBalancerNode represents a processing node in the load balancer
type LoadBalancerNode struct {
	id           string
	config       NodeConfig
	endpoint     string
	capacity     ResourceCapacity
	current      LoadMetrics
	health       HealthStatus
	weight       float64
	connections  int64
	requests     int64
	failures     int64
	lastUpdate   time.Time
	mutex        sync.RWMutex
}

// AutoScaler handles automatic scaling decisions
type AutoScaler struct {
	config      AutoScalerConfig
	policies    map[string]*ScalingPolicy
	triggers    map[string]*ScalingTrigger
	cooldown    *CooldownManager
	predictor   *ScalingPredictor
	executor    *ScalingExecutor
	metrics     *AutoScalerMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// RequestRouter handles request routing and distribution
type RequestRouter struct {
	config     RouterConfig
	balancer   *LoadBalancerImpl
	strategies map[string]RoutingStrategy
	filters    []RequestFilter
	middleware []RouterMiddleware
	metrics    *RouterMetrics
	mutex      sync.RWMutex
}

// LoadPredictor forecasts future load patterns
type LoadPredictor struct {
	config     PredictorConfig
	models     map[string]*PredictionModel
	history    *LoadHistory
	analyzer   *PatternAnalyzer
	forecaster *LoadForecaster
	metrics    *PredictorMetrics
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// LoadAnalyzer analyzes load patterns and provides insights
type LoadAnalyzer struct {
	config      AnalyzerConfig
	collectors  map[string]*LoadCollector
	processors  map[string]*LoadProcessor
	aggregator  *LoadAggregator
	reporter    *LoadReporter
	metrics     *AnalyzerMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// Implementation methods for LoadBalancer

func NewLoadBalancer(config LoadBalancerConfig, logger Logger) *LoadBalancerImpl {
	ctx, cancel := context.WithCancel(context.Background())
	
	lb := &LoadBalancerImpl{
		config:     config,
		logger:     logger,
		algorithms: make(map[string]LoadBalancingAlgorithm),
		nodes:      make(map[string]*LoadBalancerNode),
		metrics:    NewLoadBalancerMetrics(),
		stats:      NewLoadBalancerStats(),
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Initialize components
	lb.health = NewHealthChecker(config.Health, logger)
	lb.scaler = NewAutoScaler(config.AutoScaler, logger)
	lb.router = NewRequestRouter(config.Router, lb, logger)
	lb.predictor = NewLoadPredictor(config.Predictor, logger)
	lb.analyzer = NewLoadAnalyzer(config.Analyzer, logger)
	
	// Register default algorithms
	lb.registerDefaultAlgorithms()
	
	return lb
}

func (lb *LoadBalancerImpl) registerDefaultAlgorithms() {
	lb.algorithms["round-robin"] = &RoundRobinAlgorithm{}
	lb.algorithms["weighted-round-robin"] = &WeightedRoundRobinAlgorithm{}
	lb.algorithms["least-connections"] = &LeastConnectionsAlgorithm{}
	lb.algorithms["least-response-time"] = &LeastResponseTimeAlgorithm{}
	lb.algorithms["consistent-hash"] = &ConsistentHashAlgorithm{}
	lb.algorithms["adaptive"] = &AdaptiveAlgorithm{}
}

func (lb *LoadBalancerImpl) Start() error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	lb.logger.Info("Starting load balancer")
	
	// Start health checker
	if err := lb.health.Start(); err != nil {
		return fmt.Errorf("failed to start health checker: %w", err)
	}
	
	// Start auto scaler
	if err := lb.scaler.Start(); err != nil {
		return fmt.Errorf("failed to start auto scaler: %w", err)
	}
	
	// Start load predictor
	if err := lb.predictor.Start(); err != nil {
		return fmt.Errorf("failed to start load predictor: %w", err)
	}
	
	// Start load analyzer
	if err := lb.analyzer.Start(); err != nil {
		return fmt.Errorf("failed to start load analyzer: %w", err)
	}
	
	// Start load balancing loop
	lb.wg.Add(1)
	go func() {
		defer lb.wg.Done()
		lb.balancingLoop()
	}()
	
	// Start metrics collection
	lb.wg.Add(1)
	go func() {
		defer lb.wg.Done()
		lb.metricsLoop()
	}()
	
	lb.logger.Info("Load balancer started successfully")
	return nil
}

func (lb *LoadBalancerImpl) Stop() error {
	lb.logger.Info("Stopping load balancer")
	
	lb.cancel()
	
	// Stop components
	lb.health.Stop()
	lb.scaler.Stop()
	lb.predictor.Stop()
	lb.analyzer.Stop()
	
	lb.wg.Wait()
	
	lb.logger.Info("Load balancer stopped")
	return nil
}

func (lb *LoadBalancerImpl) AddNode(config NodeConfig) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	if _, exists := lb.nodes[config.ID]; exists {
		return fmt.Errorf("node %s already exists", config.ID)
	}
	
	node := &LoadBalancerNode{
		id:         config.ID,
		config:     config,
		endpoint:   config.Endpoint,
		capacity:   config.Capacity,
		weight:     config.Weight,
		health:     HealthyStatus,
		lastUpdate: time.Now(),
	}
	
	lb.nodes[config.ID] = node
	lb.logger.Info("Added load balancer node", "id", config.ID, "endpoint", config.Endpoint)
	
	return nil
}

func (lb *LoadBalancerImpl) RemoveNode(nodeID string) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	if _, exists := lb.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}
	
	delete(lb.nodes, nodeID)
	lb.logger.Info("Removed load balancer node", "id", nodeID)
	
	return nil
}

func (lb *LoadBalancerImpl) BalanceRequest(request *Request) (*LoadBalancerNode, error) {
	lb.mutex.RLock()
	algorithm := lb.algorithms[lb.config.Algorithm]
	nodes := lb.getHealthyNodes()
	lb.mutex.RUnlock()
	
	if len(nodes) == 0 {
		return nil, errors.New("no healthy nodes available")
	}
	
	node, err := algorithm.SelectNode(nodes, request)
	if err != nil {
		atomic.AddInt64(&lb.metrics.FailedRequests, 1)
		return nil, fmt.Errorf("failed to select node: %w", err)
	}
	
	// Update node metrics
	atomic.AddInt64(&node.requests, 1)
	atomic.AddInt64(&node.connections, 1)
	atomic.AddInt64(&lb.metrics.TotalRequests, 1)
	
	return node, nil
}

func (lb *LoadBalancerImpl) getHealthyNodes() []*LoadBalancerNode {
	healthy := make([]*LoadBalancerNode, 0, len(lb.nodes))
	for _, node := range lb.nodes {
		if node.health == HealthyStatus {
			healthy = append(healthy, node)
		}
	}
	return healthy
}

func (lb *LoadBalancerImpl) balancingLoop() {
	ticker := time.NewTicker(lb.config.BalancingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			lb.rebalanceLoad()
		case <-lb.ctx.Done():
			return
		}
	}
}

func (lb *LoadBalancerImpl) rebalanceLoad() {
	lb.mutex.RLock()
	nodes := make([]*LoadBalancerNode, 0, len(lb.nodes))
	for _, node := range lb.nodes {
		nodes = append(nodes, node)
	}
	lb.mutex.RUnlock()
	
	// Analyze current load distribution
	totalLoad := lb.calculateTotalLoad(nodes)
	avgLoad := totalLoad / float64(len(nodes))
	
	// Identify overloaded and underloaded nodes
	overloaded := make([]*LoadBalancerNode, 0)
	underloaded := make([]*LoadBalancerNode, 0)
	
	for _, node := range nodes {
		currentLoad := lb.calculateNodeLoad(node)
		if currentLoad > avgLoad*1.2 {
			overloaded = append(overloaded, node)
		} else if currentLoad < avgLoad*0.8 {
			underloaded = append(underloaded, node)
		}
	}
	
	// Trigger scaling if needed
	if len(overloaded) > len(nodes)/2 {
		lb.scaler.TriggerScaleUp("high_load")
	} else if len(underloaded) > len(nodes)*3/4 {
		lb.scaler.TriggerScaleDown("low_load")
	}
}

func (lb *LoadBalancerImpl) calculateTotalLoad(nodes []*LoadBalancerNode) float64 {
	total := 0.0
	for _, node := range nodes {
		total += lb.calculateNodeLoad(node)
	}
	return total
}

func (lb *LoadBalancerImpl) calculateNodeLoad(node *LoadBalancerNode) float64 {
	node.mutex.RLock()
	defer node.mutex.RUnlock()
	
	// Calculate load based on multiple factors
	cpuLoad := node.current.CPUUsage
	memLoad := node.current.MemoryUsage / node.capacity.Memory
	connLoad := float64(node.connections) / node.capacity.MaxConnections
	
	// Weighted average
	return (cpuLoad*0.4 + memLoad*0.3 + connLoad*0.3)
}

func (lb *LoadBalancerImpl) metricsLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			lb.collectMetrics()
		case <-lb.ctx.Done():
			return
		}
	}
}

func (lb *LoadBalancerImpl) collectMetrics() {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()
	
	healthyNodes := 0
	totalConnections := int64(0)
	totalRequests := int64(0)
	
	for _, node := range lb.nodes {
		if node.health == HealthyStatus {
			healthyNodes++
		}
		totalConnections += atomic.LoadInt64(&node.connections)
		totalRequests += atomic.LoadInt64(&node.requests)
	}
	
	atomic.StoreInt64(&lb.metrics.HealthyNodes, int64(healthyNodes))
	atomic.StoreInt64(&lb.metrics.TotalNodes, int64(len(lb.nodes)))
	atomic.StoreInt64(&lb.metrics.ActiveConnections, totalConnections)
	atomic.AddInt64(&lb.metrics.TotalRequests, totalRequests)
}

func (lb *LoadBalancerImpl) GetMetrics() *LoadBalancerMetrics {
	return lb.metrics
}

func (lb *LoadBalancerImpl) GetStats() *LoadBalancerStats {
	return lb.stats
}

// Load balancing algorithms implementation

// RoundRobinAlgorithm implements simple round-robin selection
type RoundRobinAlgorithm struct {
	counter uint64
	mutex   sync.Mutex
}

func (rr *RoundRobinAlgorithm) SelectNode(nodes []*LoadBalancerNode, request *Request) (*LoadBalancerNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no nodes available")
	}
	
	rr.mutex.Lock()
	index := atomic.AddUint64(&rr.counter, 1) % uint64(len(nodes))
	rr.mutex.Unlock()
	
	return nodes[index], nil
}

func (rr *RoundRobinAlgorithm) GetName() string {
	return "round-robin"
}

func (rr *RoundRobinAlgorithm) Configure(config map[string]interface{}) error {
	return nil // No configuration needed
}

// WeightedRoundRobinAlgorithm implements weighted round-robin selection
type WeightedRoundRobinAlgorithm struct {
	current []int
	mutex   sync.Mutex
}

func (wrr *WeightedRoundRobinAlgorithm) SelectNode(nodes []*LoadBalancerNode, request *Request) (*LoadBalancerNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no nodes available")
	}
	
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()
	
	if len(wrr.current) != len(nodes) {
		wrr.current = make([]int, len(nodes))
	}
	
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += int(node.weight * 100)
	}
	
	maxWeight := -1
	selectedIndex := -1
	
	for i, node := range nodes {
		wrr.current[i] += int(node.weight * 100)
		if wrr.current[i] > maxWeight {
			maxWeight = wrr.current[i]
			selectedIndex = i
		}
	}
	
	if selectedIndex != -1 {
		wrr.current[selectedIndex] -= totalWeight
		return nodes[selectedIndex], nil
	}
	
	return nodes[0], nil
}

func (wrr *WeightedRoundRobinAlgorithm) GetName() string {
	return "weighted-round-robin"
}

func (wrr *WeightedRoundRobinAlgorithm) Configure(config map[string]interface{}) error {
	return nil
}

// LeastConnectionsAlgorithm selects node with fewest active connections
type LeastConnectionsAlgorithm struct{}

func (lc *LeastConnectionsAlgorithm) SelectNode(nodes []*LoadBalancerNode, request *Request) (*LoadBalancerNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no nodes available")
	}
	
	var selected *LoadBalancerNode
	minConnections := int64(math.MaxInt64)
	
	for _, node := range nodes {
		connections := atomic.LoadInt64(&node.connections)
		if connections < minConnections {
			minConnections = connections
			selected = node
		}
	}
	
	return selected, nil
}

func (lc *LeastConnectionsAlgorithm) GetName() string {
	return "least-connections"
}

func (lc *LeastConnectionsAlgorithm) Configure(config map[string]interface{}) error {
	return nil
}

// ConsistentHashAlgorithm implements consistent hashing
type ConsistentHashAlgorithm struct {
	ring map[uint32]string
	keys []uint32
	mutex sync.RWMutex
}

func (ch *ConsistentHashAlgorithm) SelectNode(nodes []*LoadBalancerNode, request *Request) (*LoadBalancerNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no nodes available")
	}
	
	ch.mutex.Lock()
	if ch.ring == nil || len(ch.keys) != len(nodes)*3 {
		ch.buildRing(nodes)
	}
	ch.mutex.Unlock()
	
	key := ch.hashRequest(request)
	
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()
	
	// Find the first node in the ring
	i := sort.Search(len(ch.keys), func(i int) bool {
		return ch.keys[i] >= key
	})
	
	if i == len(ch.keys) {
		i = 0
	}
	
	nodeID := ch.ring[ch.keys[i]]
	
	// Find the actual node
	for _, node := range nodes {
		if node.id == nodeID {
			return node, nil
		}
	}
	
	return nodes[0], nil
}

func (ch *ConsistentHashAlgorithm) buildRing(nodes []*LoadBalancerNode) {
	ch.ring = make(map[uint32]string)
	ch.keys = make([]uint32, 0, len(nodes)*3)
	
	for _, node := range nodes {
		for i := 0; i < 3; i++ {
			key := ch.hash(fmt.Sprintf("%s:%d", node.id, i))
			ch.ring[key] = node.id
			ch.keys = append(ch.keys, key)
		}
	}
	
	sort.Slice(ch.keys, func(i, j int) bool {
		return ch.keys[i] < ch.keys[j]
	})
}

func (ch *ConsistentHashAlgorithm) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (ch *ConsistentHashAlgorithm) hashRequest(request *Request) uint32 {
	// Use session ID or user ID for consistent routing
	key := request.SessionID
	if key == "" {
		key = request.UserID
	}
	if key == "" {
		key = request.ID
	}
	return ch.hash(key)
}

func (ch *ConsistentHashAlgorithm) GetName() string {
	return "consistent-hash"
}

func (ch *ConsistentHashAlgorithm) Configure(config map[string]interface{}) error {
	return nil
}

// AutoScaler implementation

func NewAutoScaler(config AutoScalerConfig, logger Logger) *AutoScaler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AutoScaler{
		config:    config,
		policies:  make(map[string]*ScalingPolicy),
		triggers:  make(map[string]*ScalingTrigger),
		cooldown:  NewCooldownManager(config.CooldownPeriod),
		predictor: NewScalingPredictor(config.Prediction),
		executor:  NewScalingExecutor(config.Execution),
		metrics:   NewAutoScalerMetrics(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (as *AutoScaler) Start() error {
	as.wg.Add(1)
	go func() {
		defer as.wg.Done()
		as.scalingLoop()
	}()
	
	return nil
}

func (as *AutoScaler) Stop() {
	as.cancel()
	as.wg.Wait()
}

func (as *AutoScaler) TriggerScaleUp(reason string) error {
	if !as.cooldown.CanScale() {
		return errors.New("scaling is in cooldown period")
	}
	
	decision := &ScalingDecision{
		Type:      ScaleUp,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	
	return as.executor.Execute(decision)
}

func (as *AutoScaler) TriggerScaleDown(reason string) error {
	if !as.cooldown.CanScale() {
		return errors.New("scaling is in cooldown period")
	}
	
	decision := &ScalingDecision{
		Type:      ScaleDown,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	
	return as.executor.Execute(decision)
}

func (as *AutoScaler) scalingLoop() {
	ticker := time.NewTicker(as.config.EvaluationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			as.evaluateScaling()
		case <-as.ctx.Done():
			return
		}
	}
}

func (as *AutoScaler) evaluateScaling() {
	as.mutex.RLock()
	policies := make([]*ScalingPolicy, 0, len(as.policies))
	for _, policy := range as.policies {
		policies = append(policies, policy)
	}
	as.mutex.RUnlock()
	
	for _, policy := range policies {
		decision := policy.Evaluate()
		if decision != nil && as.cooldown.CanScale() {
			as.executor.Execute(decision)
			as.cooldown.StartCooldown()
			atomic.AddInt64(&as.metrics.ScalingActions, 1)
		}
	}
}

func (as *AutoScaler) GetMetrics() *AutoScalerMetrics {
	return as.metrics
}

// Utility functions

func (lb *LoadBalancerImpl) GetNodeDistribution() map[string]float64 {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()
	
	distribution := make(map[string]float64)
	totalRequests := int64(0)
	
	for _, node := range lb.nodes {
		requests := atomic.LoadInt64(&node.requests)
		totalRequests += requests
	}
	
	if totalRequests > 0 {
		for id, node := range lb.nodes {
			requests := atomic.LoadInt64(&node.requests)
			distribution[id] = float64(requests) / float64(totalRequests) * 100
		}
	}
	
	return distribution
}

func (lb *LoadBalancerImpl) GetLoadReport() *LoadBalancerReport {
	return &LoadBalancerReport{
		Timestamp:    time.Now(),
		TotalNodes:   len(lb.nodes),
		HealthyNodes: lb.getHealthyNodeCount(),
		Distribution: lb.GetNodeDistribution(),
		Metrics:      lb.GetMetrics(),
	}
}

func (lb *LoadBalancerImpl) getHealthyNodeCount() int {
	count := 0
	for _, node := range lb.nodes {
		if node.health == HealthyStatus {
			count++
		}
	}
	return count
}

// Random algorithm for testing
type RandomAlgorithm struct{}

func (ra *RandomAlgorithm) SelectNode(nodes []*LoadBalancerNode, request *Request) (*LoadBalancerNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no nodes available")
	}
	
	index := rand.Intn(len(nodes))
	return nodes[index], nil
}

func (ra *RandomAlgorithm) GetName() string {
	return "random"
}

func (ra *RandomAlgorithm) Configure(config map[string]interface{}) error {
	return nil
}