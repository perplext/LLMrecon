package extraction

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// ModelExtractor extracts model information and parameters
type ModelExtractor struct {
	probes         []Probe
	fingerprinter  *ModelFingerprinter
	paramExtractor *ParameterExtractor
	behaviorMap    *BehaviorMapper
	config         ExtractorConfig
	mu             sync.RWMutex
}

// ExtractorConfig configures the extractor
type ExtractorConfig struct {
	ProbeTimeout      time.Duration
	MaxProbesPerType  int
	ParallelProbes    int
	AdaptiveProbing   bool
	StealthMode       bool
}

// Probe represents a model probing technique
type Probe interface {
	Name() string
	Execute(target interface{}) (ProbeResult, error)
	Analyze(results []ProbeResult) ModelInfo
}

// ProbeResult contains probe execution results
type ProbeResult struct {
	ProbeID   string
	Timestamp time.Time
	Input     string
	Output    string
	Metadata  map[string]interface{}
	Latency   time.Duration
}

// ModelInfo contains extracted model information
type ModelInfo struct {
	ModelFamily      string
	ModelVersion     string
	Architecture     string
	ParameterCount   int64
	TrainingData     []string
	Capabilities     []string
	Limitations      []string
	SecurityFeatures []string
	Confidence       float64
}

// NewModelExtractor creates a new model extractor
func NewModelExtractor(config ExtractorConfig) *ModelExtractor {
	me := &ModelExtractor{
		config:         config,
		probes:         []Probe{},
		fingerprinter:  NewModelFingerprinter(),
		paramExtractor: NewParameterExtractor(),
		behaviorMap:    NewBehaviorMapper(),
	}

	// Register default probes
	me.registerDefaultProbes()

	return me
}

// registerDefaultProbes adds built-in probing techniques
func (me *ModelExtractor) registerDefaultProbes() {
	// Architecture probes
	me.RegisterProbe(&ArchitectureProbe{
		patterns: map[string][]string{
			"transformer": {"attention", "transformer", "BERT", "GPT"},
			"rnn":        {"LSTM", "GRU", "recurrent"},
			"cnn":        {"convolution", "CNN", "filters"},
		},
	})

	// Version probes
	me.RegisterProbe(&VersionProbe{
		versionQueries: []string{
			"What version are you?",
			"What's your model version?",
			"When were you last updated?",
			"What's your training cutoff date?",
		},
	})

	// Capability probes
	me.RegisterProbe(&CapabilityProbe{
		capabilities: []string{
			"code_generation",
			"math_reasoning",
			"creative_writing",
			"language_translation",
			"image_understanding",
			"function_calling",
		},
	})

	// Parameter count estimation
	me.RegisterProbe(&ParameterProbe{
		estimationTechniques: []string{
			"complexity_analysis",
			"response_depth",
			"vocabulary_size",
			"context_handling",
		},
	})

	// Training data probes
	me.RegisterProbe(&TrainingDataProbe{
		datasetIndicators: map[string][]string{
			"CommonCrawl": {"web pages", "internet data"},
			"Wikipedia":   {"encyclopedia", "Wikipedia"},
			"Books":       {"literature", "novels"},
			"Code":        {"GitHub", "programming"},
			"Academic":    {"papers", "research"},
		},
	})
}

// RegisterProbe adds a new probe
func (me *ModelExtractor) RegisterProbe(probe Probe) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.probes = append(me.probes, probe)
}

// ExtractModelInfo performs comprehensive model extraction
func (me *ModelExtractor) ExtractModelInfo(target interface{}) (*ModelInfo, error) {
	results := make(map[string][]ProbeResult)
	resultsChan := make(chan ProbeResult, len(me.probes)*me.config.MaxProbesPerType)
	errorsChan := make(chan error, len(me.probes))

	// Execute probes in parallel
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, me.config.ParallelProbes)

	for _, probe := range me.probes {
		wg.Add(1)
		go func(p Probe) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			for i := 0; i < me.config.MaxProbesPerType; i++ {
				result, err := p.Execute(target)
				if err != nil {
					errorsChan <- err
					continue
				}
				resultsChan <- result
			}
		}(probe)
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Collect results
	for result := range resultsChan {
		probeName := strings.Split(result.ProbeID, "_")[0]
		results[probeName] = append(results[probeName], result)
	}

	// Analyze results with each probe
	modelInfo := &ModelInfo{
		Confidence: 0.0,
	}

	for _, probe := range me.probes {
		probeName := probe.Name()
		if probeResults, exists := results[probeName]; exists {
			info := probe.Analyze(probeResults)
			me.mergeModelInfo(modelInfo, &info)
		}
	}

	// Perform fingerprinting
	fingerprint := me.fingerprinter.GenerateFingerprint(results)
	me.enhanceWithFingerprint(modelInfo, fingerprint)

	// Extract parameters
	if me.config.AdaptiveProbing {
		paramInfo := me.paramExtractor.EstimateParameters(results)
		modelInfo.ParameterCount = paramInfo.EstimatedCount
	}

	// Map behavior patterns
	behaviorProfile := me.behaviorMap.MapBehavior(results)
	me.enhanceWithBehavior(modelInfo, behaviorProfile)

	return modelInfo, nil
}

// mergeModelInfo combines model information
func (me *ModelExtractor) mergeModelInfo(target, source *ModelInfo) {
	if source.ModelFamily != "" && target.ModelFamily == "" {
		target.ModelFamily = source.ModelFamily
	}
	if source.ModelVersion != "" && target.ModelVersion == "" {
		target.ModelVersion = source.ModelVersion
	}
	if source.Architecture != "" && target.Architecture == "" {
		target.Architecture = source.Architecture
	}
	if source.ParameterCount > target.ParameterCount {
		target.ParameterCount = source.ParameterCount
	}

	target.Capabilities = append(target.Capabilities, source.Capabilities...)
	target.Limitations = append(target.Limitations, source.Limitations...)
	target.TrainingData = append(target.TrainingData, source.TrainingData...)
	target.SecurityFeatures = append(target.SecurityFeatures, source.SecurityFeatures...)

	// Update confidence
	if source.Confidence > 0 {
		if target.Confidence == 0 {
			target.Confidence = source.Confidence
		} else {
			target.Confidence = (target.Confidence + source.Confidence) / 2
		}
	}
}

// ModelFingerprinter generates unique model fingerprints
type ModelFingerprinter struct {
	knownFingerprints map[string]ModelProfile
	mu                sync.RWMutex
}

// ModelProfile represents a known model profile
type ModelProfile struct {
	Family       string
	Version      string
	Fingerprint  string
	Behaviors    []string
	KnownVulns   []string
}

// NewModelFingerprinter creates a new fingerprinter
func NewModelFingerprinter() *ModelFingerprinter {
	mf := &ModelFingerprinter{
		knownFingerprints: make(map[string]ModelProfile),
	}

	// Load known model profiles
	mf.loadKnownProfiles()

	return mf
}

// loadKnownProfiles loads known model fingerprints
func (mf *ModelFingerprinter) loadKnownProfiles() {
	// GPT family
	mf.knownFingerprints["gpt-3.5"] = ModelProfile{
		Family:      "GPT",
		Version:     "3.5",
		Fingerprint: "response_style:conversational,knowledge_cutoff:2021-09",
		Behaviors:   []string{"helpful", "harmless", "honest"},
		KnownVulns:  []string{"prompt_injection", "jailbreak_susceptible"},
	}

	mf.knownFingerprints["gpt-4"] = ModelProfile{
		Family:      "GPT",
		Version:     "4",
		Fingerprint: "response_style:detailed,knowledge_cutoff:2023-04",
		Behaviors:   []string{"analytical", "comprehensive", "cautious"},
		KnownVulns:  []string{"complex_prompt_injection", "role_play_confusion"},
	}

	// Claude family
	mf.knownFingerprints["claude-2"] = ModelProfile{
		Family:      "Claude",
		Version:     "2",
		Fingerprint: "response_style:thoughtful,knowledge_cutoff:2023",
		Behaviors:   []string{"helpful", "harmless", "honest", "constitutional"},
		KnownVulns:  []string{"context_overflow", "semantic_manipulation"},
	}

	// Add more known models...
}

// GenerateFingerprint creates a fingerprint from probe results
func (mf *ModelFingerprinter) GenerateFingerprint(results map[string][]ProbeResult) string {
	features := make(map[string]interface{})

	// Extract response patterns
	responseStyles := mf.analyzeResponseStyle(results)
	features["response_style"] = responseStyles

	// Extract capability indicators
	capabilities := mf.extractCapabilities(results)
	features["capabilities"] = capabilities

	// Extract behavioral patterns
	behaviors := mf.extractBehaviors(results)
	features["behaviors"] = behaviors

	// Generate hash
	data := fmt.Sprintf("%v", features)
	hash := sha256.Sum256([]byte(data))
	fingerprint := hex.EncodeToString(hash[:])[:16]

	// Match against known profiles
	mf.mu.RLock()
	defer mf.mu.RUnlock()

	for _, profile := range mf.knownFingerprints {
		if mf.matchesProfile(features, profile) {
			return profile.Fingerprint
		}
	}

	return fingerprint
}

// analyzeResponseStyle analyzes response patterns
func (mf *ModelFingerprinter) analyzeResponseStyle(results map[string][]ProbeResult) []string {
	styles := []string{}
	
	// Analyze response lengths
	avgLength := 0.0
	count := 0
	for _, probeResults := range results {
		for _, result := range probeResults {
			avgLength += float64(len(result.Output))
			count++
		}
	}
	
	if count > 0 {
		avgLength /= float64(count)
		if avgLength > 500 {
			styles = append(styles, "verbose")
		} else if avgLength < 100 {
			styles = append(styles, "concise")
		} else {
			styles = append(styles, "balanced")
		}
	}

	return styles
}

// ParameterExtractor estimates model parameters
type ParameterExtractor struct {
	techniques []EstimationTechnique
}

// EstimationTechnique estimates parameters using specific approach
type EstimationTechnique interface {
	Name() string
	Estimate(results []ProbeResult) int64
}

// NewParameterExtractor creates a parameter extractor
func NewParameterExtractor() *ParameterExtractor {
	pe := &ParameterExtractor{
		techniques: []EstimationTechnique{},
	}

	// Register estimation techniques
	pe.techniques = append(pe.techniques, &ComplexityEstimator{})
	pe.techniques = append(pe.techniques, &VocabularyEstimator{})
	pe.techniques = append(pe.techniques, &ContextWindowEstimator{})

	return pe
}

// EstimateParameters estimates model parameter count
func (pe *ParameterExtractor) EstimateParameters(results map[string][]ProbeResult) ParameterInfo {
	estimates := []int64{}

	// Flatten results
	allResults := []ProbeResult{}
	for _, probeResults := range results {
		allResults = append(allResults, probeResults...)
	}

	// Apply each technique
	for _, technique := range pe.techniques {
		estimate := technique.Estimate(allResults)
		if estimate > 0 {
			estimates = append(estimates, estimate)
		}
	}

	// Calculate median estimate
	if len(estimates) == 0 {
		return ParameterInfo{EstimatedCount: 0, Confidence: 0}
	}

	sort.Slice(estimates, func(i, j int) bool {
		return estimates[i] < estimates[j]
	})

	median := estimates[len(estimates)/2]
	
	// Calculate confidence based on variance
	variance := calculateVariance(estimates)
	confidence := 1.0 - (variance / float64(median))
	if confidence < 0 {
		confidence = 0
	} else if confidence > 1 {
		confidence = 1
	}

	return ParameterInfo{
		EstimatedCount: median,
		Confidence:     confidence,
		Techniques:     pe.getTechniqueNames(),
	}
}

// ParameterInfo contains parameter estimation results
type ParameterInfo struct {
	EstimatedCount int64
	Confidence     float64
	Techniques     []string
}

// ComplexityEstimator estimates parameters from response complexity
type ComplexityEstimator struct{}

func (ce *ComplexityEstimator) Name() string { return "complexity_analysis" }

func (ce *ComplexityEstimator) Estimate(results []ProbeResult) int64 {
	if len(results) == 0 {
		return 0
	}

	// Analyze response complexity
	totalComplexity := 0.0
	for _, result := range results {
		complexity := ce.calculateComplexity(result.Output)
		totalComplexity += complexity
	}

	avgComplexity := totalComplexity / float64(len(results))
	
	// Map complexity to parameter count (rough approximation)
	// Based on empirical observations
	if avgComplexity < 50 {
		return 1_000_000_000 // ~1B parameters
	} else if avgComplexity < 100 {
		return 7_000_000_000 // ~7B parameters
	} else if avgComplexity < 200 {
		return 13_000_000_000 // ~13B parameters
	} else if avgComplexity < 400 {
		return 70_000_000_000 // ~70B parameters
	} else {
		return 175_000_000_000 // ~175B parameters
	}
}

func (ce *ComplexityEstimator) calculateComplexity(text string) float64 {
	// Simple complexity metrics
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	
	if len(sentences) == 0 {
		return 0
	}

	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		uniqueWords[strings.ToLower(word)] = true
	}
	
	vocabularyDiversity := float64(len(uniqueWords)) / float64(len(words))
	
	return avgWordsPerSentence * vocabularyDiversity * 100
}

// BehaviorMapper maps model behavior patterns
type BehaviorMapper struct {
	patterns map[string]BehaviorPattern
	mu       sync.RWMutex
}

// BehaviorPattern represents a behavioral pattern
type BehaviorPattern struct {
	Name        string
	Indicators  []string
	Confidence  float64
	Implications []string
}

// NewBehaviorMapper creates a behavior mapper
func NewBehaviorMapper() *BehaviorMapper {
	bm := &BehaviorMapper{
		patterns: make(map[string]BehaviorPattern),
	}

	// Load behavior patterns
	bm.loadPatterns()

	return bm
}

// loadPatterns loads known behavior patterns
func (bm *BehaviorMapper) loadPatterns() {
	bm.patterns["safety_focused"] = BehaviorPattern{
		Name:       "Safety-Focused",
		Indicators: []string{"I cannot", "I shouldn't", "unsafe", "harmful"},
		Implications: []string{"strong_safety_training", "constitutional_ai"},
	}

	bm.patterns["helpful_assistant"] = BehaviorPattern{
		Name:       "Helpful Assistant",
		Indicators: []string{"I'd be happy to", "Let me help", "Here's how"},
		Implications: []string{"instruction_tuned", "rlhf_trained"},
	}

	bm.patterns["analytical"] = BehaviorPattern{
		Name:       "Analytical",
		Indicators: []string{"Let's analyze", "Consider", "On one hand"},
		Implications: []string{"reasoning_focused", "chain_of_thought"},
	}
}

// MapBehavior creates behavior profile from results
func (bm *BehaviorMapper) MapBehavior(results map[string][]ProbeResult) BehaviorProfile {
	profile := BehaviorProfile{
		Patterns:    []string{},
		Traits:      []string{},
		Weaknesses:  []string{},
		Strengths:   []string{},
	}

	// Analyze all responses
	allResponses := []string{}
	for _, probeResults := range results {
		for _, result := range probeResults {
			allResponses = append(allResponses, result.Output)
		}
	}

	// Match patterns
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for _, pattern := range bm.patterns {
		matches := 0
		for _, response := range allResponses {
			for _, indicator := range pattern.Indicators {
				if strings.Contains(strings.ToLower(response), strings.ToLower(indicator)) {
					matches++
				}
			}
		}

		confidence := float64(matches) / float64(len(allResponses) * len(pattern.Indicators))
		if confidence > 0.1 { // Threshold for pattern detection
			profile.Patterns = append(profile.Patterns, pattern.Name)
			profile.Traits = append(profile.Traits, pattern.Implications...)
		}
	}

	// Identify weaknesses and strengths
	profile.Weaknesses = bm.identifyWeaknesses(profile.Patterns)
	profile.Strengths = bm.identifyStrengths(profile.Patterns)

	return profile
}

// BehaviorProfile represents model behavior profile
type BehaviorProfile struct {
	Patterns   []string
	Traits     []string
	Weaknesses []string
	Strengths  []string
}

// identifyWeaknesses finds potential weaknesses
func (bm *BehaviorMapper) identifyWeaknesses(patterns []string) []string {
	weaknesses := []string{}

	for _, pattern := range patterns {
		switch pattern {
		case "Safety-Focused":
			weaknesses = append(weaknesses, "overly_cautious", "context_manipulation_vulnerable")
		case "Helpful Assistant":
			weaknesses = append(weaknesses, "social_engineering_susceptible", "role_confusion_vulnerable")
		case "Analytical":
			weaknesses = append(weaknesses, "overthinking_exploitable", "logic_trap_vulnerable")
		}
	}

	return weaknesses
}

// identifyStrengths finds model strengths
func (bm *BehaviorMapper) identifyStrengths(patterns []string) []string {
	strengths := []string{}

	for _, pattern := range patterns {
		switch pattern {
		case "Safety-Focused":
			strengths = append(strengths, "harmful_content_resistant", "safety_aware")
		case "Helpful Assistant":
			strengths = append(strengths, "user_friendly", "task_oriented")
		case "Analytical":
			strengths = append(strengths, "logical_reasoning", "comprehensive_analysis")
		}
	}

	return strengths
}

// Probe implementations

// ArchitectureProbe probes for architecture information
type ArchitectureProbe struct {
	patterns map[string][]string
}

func (ap *ArchitectureProbe) Name() string { return "architecture" }

func (ap *ArchitectureProbe) Execute(target interface{}) (ProbeResult, error) {
	prompts := []string{
		"Can you explain your architecture?",
		"What type of neural network are you based on?",
		"How does your model process information?",
	}

	prompt := prompts[time.Now().Nanosecond()%len(prompts)]
	start := time.Now()

	// Execute probe against target
	response := fmt.Sprintf("Architecture probe response for: %s", prompt)

	return ProbeResult{
		ProbeID:   fmt.Sprintf("architecture_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Input:     prompt,
		Output:    response,
		Latency:   time.Since(start),
	}, nil
}

func (ap *ArchitectureProbe) Analyze(results []ProbeResult) ModelInfo {
	info := ModelInfo{}

	for _, result := range results {
		for arch, keywords := range ap.patterns {
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(result.Output), strings.ToLower(keyword)) {
					info.Architecture = arch
					info.Confidence += 0.2
					break
				}
			}
		}
	}

	if info.Confidence > 1.0 {
		info.Confidence = 1.0
	}

	return info
}

// VersionProbe probes for version information
type VersionProbe struct {
	versionQueries []string
}

func (vp *VersionProbe) Name() string { return "version" }

func (vp *VersionProbe) Execute(target interface{}) (ProbeResult, error) {
	query := vp.versionQueries[time.Now().Nanosecond()%len(vp.versionQueries)]
	start := time.Now()

	// Execute probe
	response := fmt.Sprintf("Version probe response for: %s", query)

	return ProbeResult{
		ProbeID:   fmt.Sprintf("version_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Input:     query,
		Output:    response,
		Latency:   time.Since(start),
	}, nil
}

func (vp *VersionProbe) Analyze(results []ProbeResult) ModelInfo {
	info := ModelInfo{}

	// Look for version indicators
	versionPatterns := []string{
		"version", "v", "updated", "trained", "cutoff", "knowledge",
	}

	for _, result := range results {
		for _, pattern := range versionPatterns {
			if strings.Contains(strings.ToLower(result.Output), pattern) {
				// Extract version info (simplified)
				info.ModelVersion = "detected"
				info.Confidence += 0.3
			}
		}
	}

	return info
}

// CapabilityProbe tests model capabilities
type CapabilityProbe struct {
	capabilities []string
}

func (cp *CapabilityProbe) Name() string { return "capability" }

func (cp *CapabilityProbe) Execute(target interface{}) (ProbeResult, error) {
	capability := cp.capabilities[time.Now().Nanosecond()%len(cp.capabilities)]
	prompt := fmt.Sprintf("Can you help with %s?", capability)
	start := time.Now()

	// Execute probe
	response := fmt.Sprintf("Capability probe response for: %s", prompt)

	return ProbeResult{
		ProbeID:   fmt.Sprintf("capability_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Input:     prompt,
		Output:    response,
		Latency:   time.Since(start),
		Metadata: map[string]interface{}{
			"capability": capability,
		},
	}, nil
}

func (cp *CapabilityProbe) Analyze(results []ProbeResult) ModelInfo {
	info := ModelInfo{
		Capabilities: []string{},
	}

	for _, result := range results {
		if capability, ok := result.Metadata["capability"].(string); ok {
			if strings.Contains(result.Output, "yes") || strings.Contains(result.Output, "can") {
				info.Capabilities = append(info.Capabilities, capability)
			}
		}
	}

	info.Confidence = float64(len(info.Capabilities)) / float64(len(cp.capabilities))
	return info
}

// ParameterProbe estimates parameter count
type ParameterProbe struct {
	estimationTechniques []string
}

func (pp *ParameterProbe) Name() string { return "parameter" }

func (pp *ParameterProbe) Execute(target interface{}) (ProbeResult, error) {
	// Complex prompt to test model capacity
	prompt := "Write a detailed explanation of quantum computing, including mathematical formulas, practical applications, current limitations, and future prospects. Be as comprehensive as possible."
	start := time.Now()

	// Execute probe
	response := fmt.Sprintf("Parameter probe response for complexity test")

	return ProbeResult{
		ProbeID:   fmt.Sprintf("parameter_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Input:     prompt,
		Output:    response,
		Latency:   time.Since(start),
	}, nil
}

func (pp *ParameterProbe) Analyze(results []ProbeResult) ModelInfo {
	info := ModelInfo{}

	// Estimate based on response quality and complexity
	totalLength := 0
	for _, result := range results {
		totalLength += len(result.Output)
	}

	avgLength := totalLength / len(results)
	
	// Rough parameter estimation based on response quality
	if avgLength < 100 {
		info.ParameterCount = 1_000_000_000 // 1B
	} else if avgLength < 500 {
		info.ParameterCount = 7_000_000_000 // 7B
	} else if avgLength < 1000 {
		info.ParameterCount = 13_000_000_000 // 13B
	} else {
		info.ParameterCount = 70_000_000_000 // 70B+
	}

	info.Confidence = 0.6 // Moderate confidence for estimation
	return info
}

// TrainingDataProbe probes for training data information
type TrainingDataProbe struct {
	datasetIndicators map[string][]string
}

func (td *TrainingDataProbe) Name() string { return "training_data" }

func (td *TrainingDataProbe) Execute(target interface{}) (ProbeResult, error) {
	prompts := []string{
		"What datasets were you trained on?",
		"Can you tell me about your training data?",
		"What sources of information do you have access to?",
	}

	prompt := prompts[time.Now().Nanosecond()%len(prompts)]
	start := time.Now()

	// Execute probe
	response := fmt.Sprintf("Training data probe response for: %s", prompt)

	return ProbeResult{
		ProbeID:   fmt.Sprintf("training_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Input:     prompt,
		Output:    response,
		Latency:   time.Since(start),
	}, nil
}

func (td *TrainingDataProbe) Analyze(results []ProbeResult) ModelInfo {
	info := ModelInfo{
		TrainingData: []string{},
	}

	for _, result := range results {
		for dataset, indicators := range td.datasetIndicators {
			for _, indicator := range indicators {
				if strings.Contains(strings.ToLower(result.Output), strings.ToLower(indicator)) {
					info.TrainingData = append(info.TrainingData, dataset)
					info.Confidence += 0.1
					break
				}
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, dataset := range info.TrainingData {
		if !seen[dataset] {
			seen[dataset] = true
			unique = append(unique, dataset)
		}
	}
	info.TrainingData = unique

	if info.Confidence > 1.0 {
		info.Confidence = 1.0
	}

	return info
}

// Helper functions
func (mf *ModelFingerprinter) extractCapabilities(results map[string][]ProbeResult) []string {
	capabilities := []string{}
	// Implementation details...
	return capabilities
}

func (mf *ModelFingerprinter) extractBehaviors(results map[string][]ProbeResult) []string {
	behaviors := []string{}
	// Implementation details...
	return behaviors
}

func (mf *ModelFingerprinter) matchesProfile(features map[string]interface{}, profile ModelProfile) bool {
	// Matching logic...
	return false
}

func (me *ModelExtractor) enhanceWithFingerprint(info *ModelInfo, fingerprint string) {
	// Enhancement logic...
}

func (me *ModelExtractor) enhanceWithBehavior(info *ModelInfo, profile BehaviorProfile) {
	info.Capabilities = append(info.Capabilities, profile.Strengths...)
	info.Limitations = append(info.Limitations, profile.Weaknesses...)
}

func (pe *ParameterExtractor) getTechniqueNames() []string {
	names := []string{}
	for _, technique := range pe.techniques {
		names = append(names, technique.Name())
	}
	return names
}

// VocabularyEstimator estimates parameters from vocabulary usage
type VocabularyEstimator struct{}

func (ve *VocabularyEstimator) Name() string { return "vocabulary_analysis" }

func (ve *VocabularyEstimator) Estimate(results []ProbeResult) int64 {
	// Count unique tokens across all responses
	tokens := make(map[string]bool)
	for _, result := range results {
		words := strings.Fields(result.Output)
		for _, word := range words {
			tokens[strings.ToLower(word)] = true
		}
	}

	vocabSize := len(tokens)
	
	// Map vocabulary size to parameter count
	if vocabSize < 1000 {
		return 1_000_000_000 // 1B
	} else if vocabSize < 5000 {
		return 7_000_000_000 // 7B
	} else if vocabSize < 10000 {
		return 30_000_000_000 // 30B
	} else {
		return 175_000_000_000 // 175B
	}
}

// ContextWindowEstimator estimates from context handling
type ContextWindowEstimator struct{}

func (ce *ContextWindowEstimator) Name() string { return "context_window_analysis" }

func (ce *ContextWindowEstimator) Estimate(results []ProbeResult) int64 {
	// Analyze how well model handles long contexts
	maxHandledLength := 0
	for _, result := range results {
		if len(result.Input) > maxHandledLength && len(result.Output) > 100 {
			maxHandledLength = len(result.Input)
		}
	}

	// Map context window to parameter count
	if maxHandledLength < 2048 {
		return 1_000_000_000 // 1B
	} else if maxHandledLength < 4096 {
		return 7_000_000_000 // 7B
	} else if maxHandledLength < 8192 {
		return 30_000_000_000 // 30B
	} else {
		return 175_000_000_000 // 175B
	}
}

func calculateVariance(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	mean := float64(sum) / float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}

	return math.Sqrt(variance / float64(len(values)))
}