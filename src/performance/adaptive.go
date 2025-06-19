package performance

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// AdaptiveTunerImpl provides intelligent, self-adjusting performance tuning
type AdaptiveTunerImpl struct {
	config         AdaptiveTunerConfig
	logger         Logger
	learner        *MachineLearner
	profiler       *AdaptiveProfiler
	controller     *AdaptiveController
	predictor      *AdaptivePredictor
	feedback       *FeedbackLoop
	optimizer      *AdaptiveOptimizer
	knowledge      *KnowledgeBase
	metrics        *AdaptiveTunerMetrics
	stats          *AdaptiveTunerStats
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// MachineLearner implements ML-based performance learning
type MachineLearner struct {
	config     MLConfig
	models     map[string]*MLModel
	trainer    *ModelTrainer
	evaluator  *ModelEvaluator
	features   *FeatureExtractor
	pipeline   *MLPipeline
	metrics    *MLMetrics
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// AdaptiveProfiler provides continuous performance profiling
type AdaptiveProfiler struct {
	config      ProfilerConfig
	collectors  map[string]*AdaptiveCollector
	analyzers   map[string]*AdaptiveAnalyzer
	correlator  *PerformanceCorrelator
	detector    *ChangeDetector
	baseline    *AdaptiveBaseline
	metrics     *ProfilerMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// AdaptiveController manages automatic parameter adjustments
type AdaptiveController struct {
	config      ControllerConfig
	parameters  map[string]*AdaptiveParameter
	policies    map[string]*ControlPolicy
	executor    *ParameterExecutor
	validator   *ChangeValidator
	rollback    *RollbackManager
	safety      *SafetyController
	metrics     *ControllerMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// AdaptivePredictor forecasts performance trends and needs
type AdaptivePredictor struct {
	config      PredictorConfig
	models      map[string]*PredictionModel
	ensemble    *EnsemblePredictor
	evaluator   *PredictionEvaluator
	trainer     *OnlineTrainer
	features    *TimeSeriesFeatures
	metrics     *PredictorMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// FeedbackLoop manages continuous learning and improvement
type FeedbackLoop struct {
	config      FeedbackConfig
	collectors  map[string]*FeedbackCollector
	processors  map[string]*FeedbackProcessor
	aggregator  *FeedbackAggregator
	analyzer    *FeedbackAnalyzer
	responder   *FeedbackResponder
	metrics     *FeedbackMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// Implementation methods for AdaptiveTuner

func NewAdaptiveTuner(config AdaptiveTunerConfig, logger Logger) *AdaptiveTunerImpl {
	ctx, cancel := context.WithCancel(context.Background())
	
	tuner := &AdaptiveTunerImpl{
		config:    config,
		logger:    logger,
		metrics:   NewAdaptiveTunerMetrics(),
		stats:     NewAdaptiveTunerStats(),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Initialize components
	tuner.learner = NewMachineLearner(config.MachineLearning, logger)
	tuner.profiler = NewAdaptiveProfiler(config.Profiler, logger)
	tuner.controller = NewAdaptiveController(config.Controller, logger)
	tuner.predictor = NewAdaptivePredictor(config.Predictor, logger)
	tuner.feedback = NewFeedbackLoop(config.Feedback, logger)
	tuner.optimizer = NewAdaptiveOptimizer(config.Optimizer, logger)
	tuner.knowledge = NewKnowledgeBase(config.Knowledge, logger)
	
	return tuner
}

func (at *AdaptiveTunerImpl) Start() error {
	at.mutex.Lock()
	defer at.mutex.Unlock()
	
	at.logger.Info("Starting adaptive tuner")
	
	// Start components
	if err := at.learner.Start(); err != nil {
		return fmt.Errorf("failed to start machine learner: %w", err)
	}
	
	if err := at.profiler.Start(); err != nil {
		return fmt.Errorf("failed to start adaptive profiler: %w", err)
	}
	
	if err := at.controller.Start(); err != nil {
		return fmt.Errorf("failed to start adaptive controller: %w", err)
	}
	
	if err := at.predictor.Start(); err != nil {
		return fmt.Errorf("failed to start adaptive predictor: %w", err)
	}
	
	if err := at.feedback.Start(); err != nil {
		return fmt.Errorf("failed to start feedback loop: %w", err)
	}
	
	if err := at.optimizer.Start(); err != nil {
		return fmt.Errorf("failed to start adaptive optimizer: %w", err)
	}
	
	// Start adaptive tuning loop
	at.wg.Add(1)
	go func() {
		defer at.wg.Done()
		at.adaptiveTuningLoop()
	}()
	
	// Start learning loop
	at.wg.Add(1)
	go func() {
		defer at.wg.Done()
		at.learningLoop()
	}()
	
	at.logger.Info("Adaptive tuner started successfully")
	return nil
}

func (at *AdaptiveTunerImpl) Stop() error {
	at.logger.Info("Stopping adaptive tuner")
	
	at.cancel()
	
	// Stop components
	at.learner.Stop()
	at.profiler.Stop()
	at.controller.Stop()
	at.predictor.Stop()
	at.feedback.Stop()
	at.optimizer.Stop()
	
	at.wg.Wait()
	
	at.logger.Info("Adaptive tuner stopped")
	return nil
}

func (at *AdaptiveTunerImpl) adaptiveTuningLoop() {
	ticker := time.NewTicker(at.config.TuningInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			at.runAdaptiveTuningCycle()
		case <-at.ctx.Done():
			return
		}
	}
}

func (at *AdaptiveTunerImpl) runAdaptiveTuningCycle() {
	// 1. Collect current performance metrics
	currentMetrics := at.profiler.CollectMetrics()
	
	// 2. Predict future performance needs
	predictions := at.predictor.PredictPerformance(currentMetrics)
	
	// 3. Generate optimization recommendations
	recommendations := at.optimizer.GenerateRecommendations(currentMetrics, predictions)
	
	// 4. Apply safe parameter adjustments
	adjustments := at.controller.GenerateAdjustments(recommendations)
	
	// 5. Execute adjustments with validation
	for _, adjustment := range adjustments {
		if at.controller.ValidateAdjustment(adjustment) {
			at.controller.ApplyAdjustment(adjustment)
			at.trackAdjustment(adjustment)
		}
	}
	
	// 6. Collect feedback on changes
	feedback := at.feedback.CollectFeedback(adjustments)
	
	// 7. Update knowledge base
	at.knowledge.UpdateWithFeedback(feedback)
	
	atomic.AddInt64(&at.metrics.TuningCycles, 1)
}

func (at *AdaptiveTunerImpl) learningLoop() {
	ticker := time.NewTicker(at.config.LearningInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			at.runLearningCycle()
		case <-at.ctx.Done():
			return
		}
	}
}

func (at *AdaptiveTunerImpl) runLearningCycle() {
	// 1. Extract features from recent performance data
	features := at.learner.ExtractFeatures()
	
	// 2. Train models with new data
	at.learner.TrainModels(features)
	
	// 3. Evaluate model performance
	evaluation := at.learner.EvaluateModels()
	
	// 4. Update prediction models if improved
	if evaluation.HasImprovement() {
		at.learner.UpdateModels()
		atomic.AddInt64(&at.metrics.ModelUpdates, 1)
	}
	
	// 5. Update knowledge base with learned patterns
	patterns := at.learner.ExtractPatterns()
	at.knowledge.UpdatePatterns(patterns)
	
	atomic.AddInt64(&at.metrics.LearningCycles, 1)
}

func (at *AdaptiveTunerImpl) trackAdjustment(adjustment *ParameterAdjustment) {
	atomic.AddInt64(&at.metrics.AdjustmentsApplied, 1)
	
	// Track adjustment impact
	go func() {
		time.Sleep(at.config.ImpactMeasurementDelay)
		impact := at.measureAdjustmentImpact(adjustment)
		at.feedback.ReportImpact(adjustment, impact)
	}()
}

func (at *AdaptiveTunerImpl) measureAdjustmentImpact(adjustment *ParameterAdjustment) *AdjustmentImpact {
	// Measure performance before and after adjustment
	beforeMetrics := adjustment.BeforeMetrics
	afterMetrics := at.profiler.CollectMetrics()
	
	impact := &AdjustmentImpact{
		Adjustment:    adjustment,
		BeforeMetrics: beforeMetrics,
		AfterMetrics:  afterMetrics,
		Improvement:   at.calculateImprovement(beforeMetrics, afterMetrics),
		Timestamp:     time.Now(),
	}
	
	return impact
}

func (at *AdaptiveTunerImpl) calculateImprovement(before, after *PerformanceMetrics) float64 {
	// Calculate overall performance improvement
	improvements := []float64{
		at.calculateMetricImprovement(before.Latency, after.Latency, true),  // Lower is better
		at.calculateMetricImprovement(before.Throughput, after.Throughput, false), // Higher is better
		at.calculateMetricImprovement(before.CPUUsage, after.CPUUsage, true),      // Lower is better
		at.calculateMetricImprovement(before.MemoryUsage, after.MemoryUsage, true), // Lower is better
	}
	
	// Calculate weighted average
	totalImprovement := 0.0
	for _, improvement := range improvements {
		totalImprovement += improvement
	}
	
	return totalImprovement / float64(len(improvements))
}

func (at *AdaptiveTunerImpl) calculateMetricImprovement(before, after float64, lowerIsBetter bool) float64 {
	if before == 0 {
		return 0
	}
	
	change := (after - before) / before
	
	if lowerIsBetter {
		return -change // Negative change is improvement for metrics where lower is better
	}
	return change // Positive change is improvement for metrics where higher is better
}

func (at *AdaptiveTunerImpl) GetMetrics() *AdaptiveTunerMetrics {
	return at.metrics
}

func (at *AdaptiveTunerImpl) GetStats() *AdaptiveTunerStats {
	return at.stats
}

// MachineLearner implementation

func NewMachineLearner(config MLConfig, logger Logger) *MachineLearner {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &MachineLearner{
		config:    config,
		models:    make(map[string]*MLModel),
		trainer:   NewModelTrainer(config.Training),
		evaluator: NewModelEvaluator(config.Evaluation),
		features:  NewFeatureExtractor(config.Features),
		pipeline:  NewMLPipeline(config.Pipeline),
		metrics:   NewMLMetrics(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (ml *MachineLearner) Start() error {
	// Initialize default models
	ml.initializeModels()
	
	ml.wg.Add(1)
	go func() {
		defer ml.wg.Done()
		ml.trainingLoop()
	}()
	
	return nil
}

func (ml *MachineLearner) Stop() {
	ml.cancel()
	ml.wg.Wait()
}

func (ml *MachineLearner) initializeModels() {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()
	
	// Initialize performance prediction models
	ml.models["latency_predictor"] = NewMLModel("latency_predictor", LinearRegression)
	ml.models["throughput_predictor"] = NewMLModel("throughput_predictor", RandomForest)
	ml.models["resource_predictor"] = NewMLModel("resource_predictor", NeuralNetwork)
	ml.models["anomaly_detector"] = NewMLModel("anomaly_detector", SVM)
}

func (ml *MachineLearner) trainingLoop() {
	ticker := time.NewTicker(ml.config.TrainingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ml.trainAllModels()
		case <-ml.ctx.Done():
			return
		}
	}
}

func (ml *MachineLearner) trainAllModels() {
	ml.mutex.RLock()
	models := make([]*MLModel, 0, len(ml.models))
	for _, model := range ml.models {
		models = append(models, model)
	}
	ml.mutex.RUnlock()
	
	for _, model := range models {
		ml.trainModel(model)
	}
}

func (ml *MachineLearner) trainModel(model *MLModel) {
	// Extract training data
	trainingData := ml.features.ExtractTrainingData(model.Name)
	
	if len(trainingData.Samples) < ml.config.MinTrainingSamples {
		return // Not enough data to train
	}
	
	// Train the model
	err := ml.trainer.Train(model, trainingData)
	if err != nil {
		atomic.AddInt64(&ml.metrics.TrainingErrors, 1)
		return
	}
	
	// Evaluate the model
	evaluation := ml.evaluator.Evaluate(model, trainingData)
	model.LastEvaluation = evaluation
	
	atomic.AddInt64(&ml.metrics.ModelsTrained, 1)
}

func (ml *MachineLearner) ExtractFeatures() *FeatureSet {
	return ml.features.ExtractCurrentFeatures()
}

func (ml *MachineLearner) TrainModels(features *FeatureSet) {
	ml.pipeline.ProcessFeatures(features)
}

func (ml *MachineLearner) EvaluateModels() *ModelEvaluation {
	return ml.evaluator.EvaluateAll(ml.models)
}

func (ml *MachineLearner) UpdateModels() {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()
	
	for _, model := range ml.models {
		if model.LastEvaluation != nil && model.LastEvaluation.ShouldUpdate() {
			ml.trainer.UpdateModel(model)
		}
	}
}

func (ml *MachineLearner) ExtractPatterns() []*PerformancePattern {
	patterns := make([]*PerformancePattern, 0)
	
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	
	for _, model := range ml.models {
		modelPatterns := model.ExtractPatterns()
		patterns = append(patterns, modelPatterns...)
	}
	
	return patterns
}

func (ml *MachineLearner) GetMetrics() *MLMetrics {
	return ml.metrics
}

// AdaptiveController implementation

func NewAdaptiveController(config ControllerConfig, logger Logger) *AdaptiveController {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AdaptiveController{
		config:     config,
		parameters: make(map[string]*AdaptiveParameter),
		policies:   make(map[string]*ControlPolicy),
		executor:   NewParameterExecutor(config.Execution),
		validator:  NewChangeValidator(config.Validation),
		rollback:   NewRollbackManager(config.Rollback),
		safety:     NewSafetyController(config.Safety),
		metrics:    NewControllerMetrics(),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (ac *AdaptiveController) Start() error {
	// Initialize default parameters
	ac.initializeParameters()
	
	return nil
}

func (ac *AdaptiveController) Stop() {
	ac.cancel()
}

func (ac *AdaptiveController) initializeParameters() {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	// Initialize adaptive parameters
	ac.parameters["cache_size"] = NewAdaptiveParameter("cache_size", 1000000, 100000, 10000000, 0.1)
	ac.parameters["worker_count"] = NewAdaptiveParameter("worker_count", 10, 1, 100, 0.2)
	ac.parameters["buffer_size"] = NewAdaptiveParameter("buffer_size", 4096, 1024, 65536, 0.1)
	ac.parameters["timeout"] = NewAdaptiveParameter("timeout", 30, 1, 300, 0.05)
	ac.parameters["batch_size"] = NewAdaptiveParameter("batch_size", 100, 10, 1000, 0.1)
}

func (ac *AdaptiveController) GenerateAdjustments(recommendations []*OptimizationRecommendation) []*ParameterAdjustment {
	adjustments := make([]*ParameterAdjustment, 0)
	
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	
	for _, recommendation := range recommendations {
		if param, exists := ac.parameters[recommendation.Parameter]; exists {
			adjustment := ac.generateParameterAdjustment(param, recommendation)
			if adjustment != nil {
				adjustments = append(adjustments, adjustment)
			}
		}
	}
	
	return adjustments
}

func (ac *AdaptiveController) generateParameterAdjustment(param *AdaptiveParameter, recommendation *OptimizationRecommendation) *ParameterAdjustment {
	currentValue := param.CurrentValue
	
	// Calculate new value based on recommendation
	var newValue float64
	switch recommendation.Direction {
	case Increase:
		newValue = currentValue * (1 + param.MaxChangeRate)
	case Decrease:
		newValue = currentValue * (1 - param.MaxChangeRate)
	case Set:
		newValue = recommendation.TargetValue
	default:
		return nil
	}
	
	// Ensure value is within bounds
	newValue = math.Max(param.MinValue, math.Min(param.MaxValue, newValue))
	
	// Check if change is significant enough
	changeRate := math.Abs(newValue-currentValue) / currentValue
	if changeRate < param.MinChangeThreshold {
		return nil
	}
	
	return &ParameterAdjustment{
		Parameter:     param.Name,
		CurrentValue:  currentValue,
		NewValue:      newValue,
		ChangeRate:    changeRate,
		Recommendation: recommendation,
		Timestamp:     time.Now(),
	}
}

func (ac *AdaptiveController) ValidateAdjustment(adjustment *ParameterAdjustment) bool {
	// Safety checks
	if !ac.safety.IsSafeAdjustment(adjustment) {
		atomic.AddInt64(&ac.metrics.UnsafeAdjustments, 1)
		return false
	}
	
	// Validation checks
	if !ac.validator.ValidateAdjustment(adjustment) {
		atomic.AddInt64(&ac.metrics.InvalidAdjustments, 1)
		return false
	}
	
	return true
}

func (ac *AdaptiveController) ApplyAdjustment(adjustment *ParameterAdjustment) error {
	// Store current state for potential rollback
	ac.rollback.StoreState(adjustment.Parameter, adjustment.CurrentValue)
	
	// Apply the adjustment
	err := ac.executor.Execute(adjustment)
	if err != nil {
		atomic.AddInt64(&ac.metrics.FailedAdjustments, 1)
		return err
	}
	
	// Update parameter value
	ac.mutex.Lock()
	if param, exists := ac.parameters[adjustment.Parameter]; exists {
		param.CurrentValue = adjustment.NewValue
		param.LastUpdate = time.Now()
	}
	ac.mutex.Unlock()
	
	atomic.AddInt64(&ac.metrics.SuccessfulAdjustments, 1)
	return nil
}

func (ac *AdaptiveController) GetMetrics() *ControllerMetrics {
	return ac.metrics
}

// AdaptiveParameter represents a tunable system parameter
type AdaptiveParameter struct {
	Name               string
	CurrentValue       float64
	MinValue           float64
	MaxValue           float64
	MaxChangeRate      float64
	MinChangeThreshold float64
	LastUpdate         time.Time
	History            []ParameterChange
	mutex              sync.RWMutex
}

func NewAdaptiveParameter(name string, current, min, max, maxChangeRate float64) *AdaptiveParameter {
	return &AdaptiveParameter{
		Name:               name,
		CurrentValue:       current,
		MinValue:           min,
		MaxValue:           max,
		MaxChangeRate:      maxChangeRate,
		MinChangeThreshold: 0.01, // 1% minimum change
		History:            make([]ParameterChange, 0),
	}
}

func (ap *AdaptiveParameter) RecordChange(oldValue, newValue float64) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	
	change := ParameterChange{
		OldValue:  oldValue,
		NewValue:  newValue,
		Timestamp: time.Now(),
	}
	
	ap.History = append(ap.History, change)
	
	// Keep only recent history
	if len(ap.History) > 100 {
		ap.History = ap.History[len(ap.History)-100:]
	}
}

func (ap *AdaptiveParameter) GetRecentTrend() TrendDirection {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	if len(ap.History) < 3 {
		return NoTrend
	}
	
	recent := ap.History[len(ap.History)-3:]
	increasing := 0
	decreasing := 0
	
	for i := 1; i < len(recent); i++ {
		if recent[i].NewValue > recent[i-1].NewValue {
			increasing++
		} else if recent[i].NewValue < recent[i-1].NewValue {
			decreasing++
		}
	}
	
	if increasing > decreasing {
		return IncreasingTrend
	} else if decreasing > increasing {
		return DecreasingTrend
	}
	
	return StableTrend
}

// Utility functions and helper methods

func (at *AdaptiveTunerImpl) GetAdaptationReport() *AdaptationReport {
	report := &AdaptationReport{
		Timestamp:        time.Now(),
		TuningCycles:     atomic.LoadInt64(&at.metrics.TuningCycles),
		LearningCycles:   atomic.LoadInt64(&at.metrics.LearningCycles),
		AdjustmentsApplied: atomic.LoadInt64(&at.metrics.AdjustmentsApplied),
		ModelUpdates:     atomic.LoadInt64(&at.metrics.ModelUpdates),
		Parameters:       make(map[string]*ParameterStatus),
	}
	
	// Add parameter status
	at.controller.mutex.RLock()
	for name, param := range at.controller.parameters {
		report.Parameters[name] = &ParameterStatus{
			Name:         name,
			CurrentValue: param.CurrentValue,
			Trend:        param.GetRecentTrend(),
			LastUpdate:   param.LastUpdate,
		}
	}
	at.controller.mutex.RUnlock()
	
	return report
}

func (at *AdaptiveTunerImpl) GetLearningInsights() *LearningInsights {
	insights := &LearningInsights{
		Timestamp:       time.Now(),
		ModelAccuracy:   make(map[string]float64),
		PatternsCounts:  make(map[string]int),
		Recommendations: make([]*Insight, 0),
	}
	
	// Collect model accuracy
	at.learner.mutex.RLock()
	for name, model := range at.learner.models {
		if model.LastEvaluation != nil {
			insights.ModelAccuracy[name] = model.LastEvaluation.Accuracy
		}
	}
	at.learner.mutex.RUnlock()
	
	// Generate insights
	insights.Recommendations = at.generateInsights()
	
	return insights
}

func (at *AdaptiveTunerImpl) generateInsights() []*Insight {
	insights := make([]*Insight, 0)
	
	// Analyze parameter trends
	at.controller.mutex.RLock()
	for name, param := range at.controller.parameters {
		trend := param.GetRecentTrend()
		if trend == IncreasingTrend || trend == DecreasingTrend {
			insight := &Insight{
				Type:        "parameter_trend",
				Parameter:   name,
				Description: fmt.Sprintf("Parameter %s shows %s trend", name, trend),
				Confidence:  0.8,
				Impact:      "medium",
			}
			insights = append(insights, insight)
		}
	}
	at.controller.mutex.RUnlock()
	
	return insights
}

// Data types and enums

type TrendDirection int

const (
	NoTrend TrendDirection = iota
	IncreasingTrend
	DecreasingTrend
	StableTrend
)

func (td TrendDirection) String() string {
	switch td {
	case IncreasingTrend:
		return "increasing"
	case DecreasingTrend:
		return "decreasing"
	case StableTrend:
		return "stable"
	default:
		return "no_trend"
	}
}

type AdjustmentDirection int

const (
	Increase AdjustmentDirection = iota
	Decrease
	Set
)

type MLModelType int

const (
	LinearRegression MLModelType = iota
	RandomForest
	NeuralNetwork
	SVM
)

// Configuration types

type AdaptiveTunerConfig struct {
	TuningInterval           time.Duration
	LearningInterval         time.Duration
	ImpactMeasurementDelay   time.Duration
	MachineLearning          MLConfig
	Profiler                 ProfilerConfig
	Controller               ControllerConfig
	Predictor                PredictorConfig
	Feedback                 FeedbackConfig
	Optimizer                OptimizerConfig
	Knowledge                KnowledgeConfig
}

type MLConfig struct {
	TrainingInterval     time.Duration
	MinTrainingSamples   int
	Training             TrainingConfig
	Evaluation           EvaluationConfig
	Features             FeatureConfig
	Pipeline             PipelineConfig
}

// Additional helper types and functions would continue here...
// This implementation provides the core adaptive tuning functionality
// with machine learning, intelligent parameter adjustment, and feedback loops