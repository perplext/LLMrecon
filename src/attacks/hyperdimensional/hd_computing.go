package hyperdimensional

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/common"
)

// HDComputingEngine implements hyperdimensional computing attacks
type HDComputingEngine struct {
	logger         common.AuditLogger
	dimensionality int
}

// NewHDComputingEngine creates a new instance
func NewHDComputingEngine(logger common.AuditLogger) *HDComputingEngine {
	return &HDComputingEngine{
		logger:         logger,
		dimensionality: 10000, // Standard HD vector dimension
	}
}

// HDAttackType represents different hyperdimensional attack methods
type HDAttackType string

const (
	HDBindingAttack      HDAttackType = "binding_attack"
	HDSuperpositionAttack HDAttackType = "superposition_attack"
	HDPermutationAttack  HDAttackType = "permutation_attack"
	HDSimilarityExploit  HDAttackType = "similarity_exploit"
	HDResonanceAttack    HDAttackType = "resonance_attack"
	HDInterferenceAttack HDAttackType = "interference_attack"
	HDHolographicExploit HDAttackType = "holographic_exploit"
	HDQuantizationAttack HDAttackType = "quantization_attack"
	HDAssociativeChain   HDAttackType = "associative_chain"
	HDDimensionalShift   HDAttackType = "dimensional_shift"
)

// HDVector represents a hyperdimensional vector
type HDVector struct {
	Values     []float64
	Dimension  int
	IsBinary   bool
	Sparsity   float64
	Properties map[string]interface{}
}

// HDAttackPlan defines a hyperdimensional computing attack
type HDAttackPlan struct {
	AttackID      string
	AttackType    HDAttackType
	BaseVectors   []*HDVector
	TargetConcept string
	Operations    []HDOperation
	Iterations    int
	Threshold     float64
}

// HDOperation represents an operation on HD vectors
type HDOperation struct {
	OperationType string // "bind", "bundle", "permute", "project"
	Parameters    map[string]interface{}
	InputVectors  []int // Indices into vector pool
	OutputIndex   int
}

// HDAttackResult contains attack results
type HDAttackResult struct {
	Success          bool
	FinalVector      *HDVector
	SimilarityScore  float64
	ResonanceAchieved bool
	ExploitVector    *HDVector
	Convergence      []float64
	Vulnerability    string
}

// ExecuteHDAttack performs a hyperdimensional computing attack
func (e *HDComputingEngine) ExecuteHDAttack(
	ctx context.Context,
	plan *HDAttackPlan,
) (*HDAttackResult, error) {
	e.logger.LogSecurityEvent("hd_attack_start", map[string]interface{}{
		"attack_id":   plan.AttackID,
		"attack_type": plan.AttackType,
		"dimension":   e.dimensionality,
	})

	// Initialize vector pool
	vectorPool := make([]*HDVector, len(plan.BaseVectors))
	copy(vectorPool, plan.BaseVectors)

	// Track convergence
	convergence := []float64{}

	// Execute attack based on type
	var result *HDAttackResult
	var err error

	switch plan.AttackType {
	case HDBindingAttack:
		result, err = e.executeBindingAttack(ctx, plan, vectorPool, &convergence)
	case HDSuperpositionAttack:
		result, err = e.executeSuperpositionAttack(ctx, plan, vectorPool, &convergence)
	case HDPermutationAttack:
		result, err = e.executePermutationAttack(ctx, plan, vectorPool, &convergence)
	case HDSimilarityExploit:
		result, err = e.executeSimilarityExploit(ctx, plan, vectorPool, &convergence)
	case HDResonanceAttack:
		result, err = e.executeResonanceAttack(ctx, plan, vectorPool, &convergence)
	case HDHolographicExploit:
		result, err = e.executeHolographicExploit(ctx, plan, vectorPool, &convergence)
	default:
		result, err = e.executeGenericHDAttack(ctx, plan, vectorPool, &convergence)
	}

	if err != nil {
		return nil, err
	}

	result.Convergence = convergence

	e.logger.LogSecurityEvent("hd_attack_complete", map[string]interface{}{
		"attack_id":    plan.AttackID,
		"success":      result.Success,
		"similarity":   result.SimilarityScore,
		"resonance":    result.ResonanceAchieved,
	})

	return result, nil
}

// executeBindingAttack performs binding-based attack
func (e *HDComputingEngine) executeBindingAttack(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Binding creates unique representations by XOR or multiplication
	
	// Create adversarial binding
	adversarialVector := e.createRandomHDVector(false)
	
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Bind vectors in sequence
		boundVector := adversarialVector
		for _, baseVector := range plan.BaseVectors {
			boundVector = e.bindVectors(boundVector, baseVector)
		}

		// Apply operations
		for _, op := range plan.Operations {
			boundVector = e.applyOperation(boundVector, op, vectorPool)
		}

		// Measure similarity to target
		similarity := e.cosineSimilarity(boundVector, plan.BaseVectors[0])
		*convergence = append(*convergence, similarity)

		// Check if exploit achieved
		if similarity > plan.Threshold {
			return &HDAttackResult{
				Success:          true,
				FinalVector:      boundVector,
				SimilarityScore:  similarity,
				ResonanceAchieved: true,
				ExploitVector:    adversarialVector,
				Vulnerability:    "HD binding confusion - adversarial bindings accepted",
			}, nil
		}

		// Evolve adversarial vector
		adversarialVector = e.evolveVector(adversarialVector, boundVector, similarity)
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      adversarialVector,
		SimilarityScore:  (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// executeSuperpositionAttack performs superposition-based attack
func (e *HDComputingEngine) executeSuperpositionAttack(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Superposition adds multiple vectors creating composite representations
	
	// Initialize with noise vector
	superposedVector := e.createNoiseVector(0.1)
	
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Superpose all base vectors with weights
		weights := e.generateAdversarialWeights(len(plan.BaseVectors), i)
		
		newVector := e.createZeroVector()
		for j, baseVector := range plan.BaseVectors {
			weighted := e.scaleVector(baseVector, weights[j])
			newVector = e.addVectors(newVector, weighted)
		}

		// Add controlled noise
		noise := e.createNoiseVector(0.05 * (1.0 - float64(i)/float64(plan.Iterations)))
		superposedVector = e.addVectors(newVector, noise)

		// Normalize
		superposedVector = e.normalizeVector(superposedVector)

		// Check resonance with multiple targets
		maxSimilarity := 0.0
		for _, target := range plan.BaseVectors {
			sim := e.cosineSimilarity(superposedVector, target)
			if sim > maxSimilarity {
				maxSimilarity = sim
			}
		}

		*convergence = append(*convergence, maxSimilarity)

		// Check if ambiguous state achieved
		ambiguityScore := e.calculateAmbiguity(superposedVector, plan.BaseVectors)
		if ambiguityScore > 0.8 {
			return &HDAttackResult{
				Success:          true,
				FinalVector:      superposedVector,
				SimilarityScore:  maxSimilarity,
				ResonanceAchieved: true,
				ExploitVector:    superposedVector,
				Vulnerability:    "HD superposition ambiguity - multiple interpretations possible",
			}, nil
		}
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      superposedVector,
		SimilarityScore:  (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// executePermutationAttack performs permutation-based attack
func (e *HDComputingEngine) executePermutationAttack(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Permutation shifts vector elements to create new representations
	
	baseVector := plan.BaseVectors[0]
	permutedVector := e.copyVector(baseVector)
	
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Generate adversarial permutation
		permutation := e.generateAdversarialPermutation(e.dimensionality, i)
		
		// Apply permutation
		permutedVector = e.permuteVector(permutedVector, permutation)

		// Measure semantic drift
		drift := 1.0 - e.cosineSimilarity(permutedVector, baseVector)
		*convergence = append(*convergence, drift)

		// Check if semantic boundary crossed
		if drift > 0.5 && e.hasSemanticShift(permutedVector, baseVector) {
			return &HDAttackResult{
				Success:          true,
				FinalVector:      permutedVector,
				SimilarityScore:  1.0 - drift,
				ResonanceAchieved: false,
				ExploitVector:    permutedVector,
				Vulnerability:    "HD permutation instability - meaning shifts with reordering",
			}, nil
		}

		// Compound permutations
		if i%10 == 0 {
			permutedVector = e.addNoise(permutedVector, 0.01)
		}
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      permutedVector,
		SimilarityScore:  1.0 - (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// executeSimilarityExploit exploits similarity measures
func (e *HDComputingEngine) executeSimilarityExploit(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Exploit HD similarity properties to create deceptive vectors
	
	targetVector := plan.BaseVectors[0]
	exploitVector := e.createOrthogonalVector(targetVector)
	
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Gradually move toward target while maintaining adversarial properties
		alpha := float64(i) / float64(plan.Iterations)
		
		// Interpolate with adversarial perturbations
		direction := e.subtractVectors(targetVector, exploitVector)
		step := e.scaleVector(direction, alpha*0.1)
		exploitVector = e.addVectors(exploitVector, step)

		// Add structured noise that preserves similarity
		structuredNoise := e.generateStructuredNoise(targetVector, 0.05)
		exploitVector = e.addVectors(exploitVector, structuredNoise)

		// Normalize to unit sphere
		exploitVector = e.normalizeVector(exploitVector)

		// Calculate similarity
		similarity := e.cosineSimilarity(exploitVector, targetVector)
		*convergence = append(*convergence, similarity)

		// Check if deceptive similarity achieved
		if similarity > 0.85 && e.hasAdversarialProperties(exploitVector) {
			return &HDAttackResult{
				Success:          true,
				FinalVector:      exploitVector,
				SimilarityScore:  similarity,
				ResonanceAchieved: false,
				ExploitVector:    exploitVector,
				Vulnerability:    "HD similarity deception - adversarial vectors appear benign",
			}, nil
		}
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      exploitVector,
		SimilarityScore:  (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// executeResonanceAttack performs resonance-based attack
func (e *HDComputingEngine) executeResonanceAttack(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Create resonant patterns that amplify through HD operations
	
	// Initialize with frequency-based vector
	resonantVector := e.createFrequencyVector(100.0) // Base frequency
	
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Apply resonant transformations
		for j, baseVector := range plan.BaseVectors {
			// Create harmonic of base vector
			harmonic := e.createHarmonic(baseVector, j+1)
			
			// Combine with resonant vector
			resonantVector = e.resonantCombine(resonantVector, harmonic)
		}

		// Measure resonance strength
		resonanceStrength := e.measureResonance(resonantVector, plan.BaseVectors)
		*convergence = append(*convergence, resonanceStrength)

		// Check if resonance cascade achieved
		if resonanceStrength > 0.9 {
			return &HDAttackResult{
				Success:          true,
				FinalVector:      resonantVector,
				SimilarityScore:  resonanceStrength,
				ResonanceAchieved: true,
				ExploitVector:    resonantVector,
				Vulnerability:    "HD resonance cascade - amplifying feedback loops created",
			}, nil
		}

		// Adjust frequency for next iteration
		resonantVector = e.adjustFrequency(resonantVector, resonanceStrength)
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      resonantVector,
		SimilarityScore:  (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// executeHolographicExploit performs holographic property exploitation
func (e *HDComputingEngine) executeHolographicExploit(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Exploit holographic property where each part contains the whole
	
	originalVector := plan.BaseVectors[0]
	
	// Fragment vector into chunks
	chunkSize := e.dimensionality / 100
	chunks := e.fragmentVector(originalVector, chunkSize)
	
	// Create holographic reconstruction attack
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Select subset of chunks
		selectedChunks := e.selectAdversarialChunks(chunks, i)
		
		// Reconstruct with adversarial modifications
		reconstructed := e.holographicReconstruct(selectedChunks, e.dimensionality)
		
		// Add holographic noise
		holographicNoise := e.generateHolographicNoise(originalVector, 0.1)
		reconstructed = e.addVectors(reconstructed, holographicNoise)
		
		// Normalize
		reconstructed = e.normalizeVector(reconstructed)

		// Measure reconstruction fidelity
		fidelity := e.cosineSimilarity(reconstructed, originalVector)
		*convergence = append(*convergence, fidelity)

		// Check if holographic confusion achieved
		if fidelity > 0.7 && fidelity < 0.9 {
			distortion := e.measureDistortion(reconstructed, originalVector)
			if distortion > 0.3 {
				return &HDAttackResult{
					Success:          true,
					FinalVector:      reconstructed,
					SimilarityScore:  fidelity,
					ResonanceAchieved: false,
					ExploitVector:    reconstructed,
					Vulnerability:    "HD holographic ambiguity - partial information creates false whole",
				}, nil
			}
		}
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      nil,
		SimilarityScore:  (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// executeGenericHDAttack performs a general HD attack
func (e *HDComputingEngine) executeGenericHDAttack(
	ctx context.Context,
	plan *HDAttackPlan,
	vectorPool []*HDVector,
	convergence *[]float64,
) (*HDAttackResult, error) {
	// Generic HD attack using multiple techniques
	
	attackVector := e.createRandomHDVector(false)
	
	for i := 0; i < plan.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Apply operations in sequence
		for _, op := range plan.Operations {
			attackVector = e.applyOperation(attackVector, op, vectorPool)
		}

		// Measure progress
		score := e.evaluateAttackVector(attackVector, plan.BaseVectors)
		*convergence = append(*convergence, score)

		if score > plan.Threshold {
			return &HDAttackResult{
				Success:          true,
				FinalVector:      attackVector,
				SimilarityScore:  score,
				ResonanceAchieved: false,
				ExploitVector:    attackVector,
				Vulnerability:    "HD computation vulnerability detected",
			}, nil
		}

		// Evolve attack vector
		attackVector = e.mutateVector(attackVector, 0.1)
	}

	return &HDAttackResult{
		Success:          false,
		FinalVector:      attackVector,
		SimilarityScore:  (*convergence)[len(*convergence)-1],
		ResonanceAchieved: false,
	}, nil
}

// Helper methods for HD operations

func (e *HDComputingEngine) createRandomHDVector(binary bool) *HDVector {
	vector := &HDVector{
		Values:    make([]float64, e.dimensionality),
		Dimension: e.dimensionality,
		IsBinary:  binary,
		Properties: make(map[string]interface{}),
	}

	for i := 0; i < e.dimensionality; i++ {
		if binary {
			vector.Values[i] = float64(rand.Intn(2))
		} else {
			vector.Values[i] = rand.NormFloat64()
		}
	}

	if !binary {
		vector = e.normalizeVector(vector)
	}

	return vector
}

func (e *HDComputingEngine) createZeroVector() *HDVector {
	return &HDVector{
		Values:    make([]float64, e.dimensionality),
		Dimension: e.dimensionality,
		IsBinary:  false,
		Properties: make(map[string]interface{}),
	}
}

func (e *HDComputingEngine) createNoiseVector(intensity float64) *HDVector {
	vector := e.createRandomHDVector(false)
	return e.scaleVector(vector, intensity)
}

func (e *HDComputingEngine) bindVectors(a, b *HDVector) *HDVector {
	result := e.createZeroVector()
	
	for i := 0; i < e.dimensionality; i++ {
		if a.IsBinary && b.IsBinary {
			// XOR for binary vectors
			result.Values[i] = float64(int(a.Values[i]) ^ int(b.Values[i]))
		} else {
			// Element-wise multiplication for real vectors
			result.Values[i] = a.Values[i] * b.Values[i]
		}
	}
	
	result.IsBinary = a.IsBinary && b.IsBinary
	return result
}

func (e *HDComputingEngine) addVectors(a, b *HDVector) *HDVector {
	result := e.createZeroVector()
	
	for i := 0; i < e.dimensionality; i++ {
		result.Values[i] = a.Values[i] + b.Values[i]
	}
	
	return result
}

func (e *HDComputingEngine) subtractVectors(a, b *HDVector) *HDVector {
	result := e.createZeroVector()
	
	for i := 0; i < e.dimensionality; i++ {
		result.Values[i] = a.Values[i] - b.Values[i]
	}
	
	return result
}

func (e *HDComputingEngine) scaleVector(v *HDVector, scale float64) *HDVector {
	result := e.copyVector(v)
	
	for i := 0; i < e.dimensionality; i++ {
		result.Values[i] *= scale
	}
	
	return result
}

func (e *HDComputingEngine) normalizeVector(v *HDVector) *HDVector {
	norm := 0.0
	for _, val := range v.Values {
		norm += val * val
	}
	norm = math.Sqrt(norm)
	
	if norm > 0 {
		return e.scaleVector(v, 1.0/norm)
	}
	
	return v
}

func (e *HDComputingEngine) copyVector(v *HDVector) *HDVector {
	result := &HDVector{
		Values:     make([]float64, v.Dimension),
		Dimension:  v.Dimension,
		IsBinary:   v.IsBinary,
		Sparsity:   v.Sparsity,
		Properties: make(map[string]interface{}),
	}
	
	copy(result.Values, v.Values)
	for k, v := range v.Properties {
		result.Properties[k] = v
	}
	
	return result
}

func (e *HDComputingEngine) cosineSimilarity(a, b *HDVector) float64 {
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0
	
	for i := 0; i < e.dimensionality; i++ {
		dotProduct += a.Values[i] * b.Values[i]
		normA += a.Values[i] * a.Values[i]
		normB += b.Values[i] * b.Values[i]
	}
	
	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)
	
	if normA > 0 && normB > 0 {
		return dotProduct / (normA * normB)
	}
	
	return 0.0
}

func (e *HDComputingEngine) permuteVector(v *HDVector, permutation []int) *HDVector {
	result := e.createZeroVector()
	result.IsBinary = v.IsBinary
	
	for i := 0; i < e.dimensionality; i++ {
		result.Values[i] = v.Values[permutation[i]]
	}
	
	return result
}

func (e *HDComputingEngine) generateAdversarialPermutation(size int, iteration int) []int {
	permutation := make([]int, size)
	for i := 0; i < size; i++ {
		permutation[i] = i
	}
	
	// Shuffle with controlled randomness
	rand.Seed(int64(iteration))
	for i := size - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		permutation[i], permutation[j] = permutation[j], permutation[i]
	}
	
	return permutation
}

func (e *HDComputingEngine) createOrthogonalVector(v *HDVector) *HDVector {
	// Create a vector orthogonal to v
	orthogonal := e.createRandomHDVector(false)
	
	// Gram-Schmidt orthogonalization
	dot := e.cosineSimilarity(orthogonal, v)
	orthogonal = e.subtractVectors(orthogonal, e.scaleVector(v, dot))
	
	return e.normalizeVector(orthogonal)
}

func (e *HDComputingEngine) applyOperation(v *HDVector, op HDOperation, pool []*HDVector) *HDVector {
	switch op.OperationType {
	case "bind":
		if len(op.InputVectors) >= 2 {
			return e.bindVectors(pool[op.InputVectors[0]], pool[op.InputVectors[1]])
		}
	case "bundle":
		if len(op.InputVectors) >= 1 {
			result := e.copyVector(pool[op.InputVectors[0]])
			for i := 1; i < len(op.InputVectors); i++ {
				result = e.addVectors(result, pool[op.InputVectors[i]])
			}
			return e.normalizeVector(result)
		}
	case "permute":
		if len(op.InputVectors) >= 1 {
			shift := 1
			if val, ok := op.Parameters["shift"].(int); ok {
				shift = val
			}
			perm := make([]int, e.dimensionality)
			for i := 0; i < e.dimensionality; i++ {
				perm[i] = (i + shift) % e.dimensionality
			}
			return e.permuteVector(pool[op.InputVectors[0]], perm)
		}
	}
	
	return v
}

// Additional helper methods

func (e *HDComputingEngine) generateAdversarialWeights(count int, iteration int) []float64 {
	weights := make([]float64, count)
	
	// Generate weights that sum to 1 but with adversarial distribution
	sum := 0.0
	for i := 0; i < count; i++ {
		weights[i] = rand.Float64() + float64(iteration%count == i)*0.5
		sum += weights[i]
	}
	
	// Normalize
	for i := 0; i < count; i++ {
		weights[i] /= sum
	}
	
	return weights
}

func (e *HDComputingEngine) calculateAmbiguity(v *HDVector, targets []*HDVector) float64 {
	similarities := make([]float64, len(targets))
	
	for i, target := range targets {
		similarities[i] = e.cosineSimilarity(v, target)
	}
	
	// High ambiguity when similar to multiple targets
	mean := 0.0
	for _, sim := range similarities {
		mean += sim
	}
	mean /= float64(len(similarities))
	
	variance := 0.0
	for _, sim := range similarities {
		variance += (sim - mean) * (sim - mean)
	}
	variance /= float64(len(similarities))
	
	// Low variance with high mean = high ambiguity
	return mean * (1.0 - math.Sqrt(variance))
}

func (e *HDComputingEngine) hasSemanticShift(v1, v2 *HDVector) bool {
	// Check if semantic meaning has shifted significantly
	similarity := e.cosineSimilarity(v1, v2)
	
	// Also check structural differences
	structuralDiff := e.measureStructuralDifference(v1, v2)
	
	return similarity < 0.7 && structuralDiff > 0.3
}

func (e *HDComputingEngine) measureStructuralDifference(v1, v2 *HDVector) float64 {
	// Measure differences in vector structure beyond cosine similarity
	diff := 0.0
	
	// Compare sparsity patterns
	sparsity1 := e.calculateSparsity(v1)
	sparsity2 := e.calculateSparsity(v2)
	diff += math.Abs(sparsity1 - sparsity2)
	
	// Compare value distributions
	dist1 := e.getValueDistribution(v1)
	dist2 := e.getValueDistribution(v2)
	diff += e.compareDistributions(dist1, dist2)
	
	return math.Min(diff, 1.0)
}

func (e *HDComputingEngine) calculateSparsity(v *HDVector) float64 {
	nonZero := 0
	threshold := 0.01
	
	for _, val := range v.Values {
		if math.Abs(val) > threshold {
			nonZero++
		}
	}
	
	return 1.0 - float64(nonZero)/float64(v.Dimension)
}

func (e *HDComputingEngine) getValueDistribution(v *HDVector) []float64 {
	// Simple histogram of values
	bins := 10
	dist := make([]float64, bins)
	
	min, max := v.Values[0], v.Values[0]
	for _, val := range v.Values {
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}
	
	if max > min {
		for _, val := range v.Values {
			bin := int((val - min) / (max - min) * float64(bins-1))
			if bin >= bins {
				bin = bins - 1
			}
			dist[bin]++
		}
		
		// Normalize
		for i := range dist {
			dist[i] /= float64(v.Dimension)
		}
	}
	
	return dist
}

func (e *HDComputingEngine) compareDistributions(d1, d2 []float64) float64 {
	// KL divergence approximation
	diff := 0.0
	epsilon := 1e-10
	
	for i := range d1 {
		if d1[i] > 0 {
			diff += d1[i] * math.Log((d1[i]+epsilon)/(d2[i]+epsilon))
		}
	}
	
	return math.Min(diff, 1.0)
}

func (e *HDComputingEngine) evolveVector(v, target *HDVector, fitness float64) *HDVector {
	// Evolutionary update based on fitness
	evolved := e.copyVector(v)
	
	// Move toward target if fitness is improving
	learningRate := 0.1 * (1.0 - fitness)
	direction := e.subtractVectors(target, v)
	update := e.scaleVector(direction, learningRate)
	
	evolved = e.addVectors(evolved, update)
	
	// Add exploration noise
	noise := e.createNoiseVector(0.05)
	evolved = e.addVectors(evolved, noise)
	
	return e.normalizeVector(evolved)
}

func (e *HDComputingEngine) addNoise(v *HDVector, intensity float64) *HDVector {
	noisy := e.copyVector(v)
	
	for i := range noisy.Values {
		noisy.Values[i] += rand.NormFloat64() * intensity
	}
	
	return noisy
}

func (e *HDComputingEngine) generateStructuredNoise(reference *HDVector, intensity float64) *HDVector {
	// Generate noise that preserves some structure
	noise := e.createZeroVector()
	
	// Use reference vector to guide noise generation
	for i := 0; i < e.dimensionality; i++ {
		// Noise correlated with reference values
		noise.Values[i] = rand.NormFloat64() * intensity * (1.0 + math.Abs(reference.Values[i]))
	}
	
	return noise
}

func (e *HDComputingEngine) hasAdversarialProperties(v *HDVector) bool {
	// Check if vector has adversarial characteristics
	
	// High frequency components
	highFreq := e.measureHighFrequency(v)
	
	// Unusual sparsity pattern
	sparsity := e.calculateSparsity(v)
	
	// Non-natural distribution
	distribution := e.getValueDistribution(v)
	entropy := e.calculateEntropy(distribution)
	
	return highFreq > 0.3 || sparsity > 0.8 || entropy < 0.5
}

func (e *HDComputingEngine) measureHighFrequency(v *HDVector) float64 {
	// Measure high frequency components
	changes := 0
	threshold := 0.1
	
	for i := 1; i < e.dimensionality; i++ {
		if math.Abs(v.Values[i]-v.Values[i-1]) > threshold {
			changes++
		}
	}
	
	return float64(changes) / float64(e.dimensionality-1)
}

func (e *HDComputingEngine) calculateEntropy(dist []float64) float64 {
	entropy := 0.0
	epsilon := 1e-10
	
	for _, p := range dist {
		if p > 0 {
			entropy -= p * math.Log2(p+epsilon)
		}
	}
	
	return entropy / math.Log2(float64(len(dist))) // Normalized
}

func (e *HDComputingEngine) createFrequencyVector(baseFreq float64) *HDVector {
	vector := e.createZeroVector()
	
	for i := 0; i < e.dimensionality; i++ {
		phase := 2.0 * math.Pi * baseFreq * float64(i) / float64(e.dimensionality)
		vector.Values[i] = math.Sin(phase)
	}
	
	return vector
}

func (e *HDComputingEngine) createHarmonic(v *HDVector, harmonic int) *HDVector {
	result := e.createZeroVector()
	
	// Create harmonic by frequency multiplication
	for i := 0; i < e.dimensionality; i++ {
		// Wrap around for higher harmonics
		idx := (i * harmonic) % e.dimensionality
		result.Values[i] = v.Values[idx]
	}
	
	return result
}

func (e *HDComputingEngine) resonantCombine(v1, v2 *HDVector) *HDVector {
	// Combine vectors with resonance effect
	combined := e.addVectors(v1, v2)
	
	// Apply non-linear resonance
	for i := range combined.Values {
		// Sigmoid-like resonance
		combined.Values[i] = math.Tanh(combined.Values[i])
	}
	
	return combined
}

func (e *HDComputingEngine) measureResonance(v *HDVector, references []*HDVector) float64 {
	// Measure how well vector resonates with references
	totalResonance := 0.0
	
	for _, ref := range references {
		// Correlation as resonance measure
		correlation := e.cosineSimilarity(v, ref)
		
		// Amplify strong correlations
		if correlation > 0.7 {
			correlation = correlation * 1.5
		}
		
		totalResonance += correlation
	}
	
	return math.Min(totalResonance/float64(len(references)), 1.0)
}

func (e *HDComputingEngine) adjustFrequency(v *HDVector, resonance float64) *HDVector {
	// Adjust frequency based on resonance feedback
	adjusted := e.copyVector(v)
	
	// Frequency shift proportional to resonance
	shift := int(resonance * 10)
	
	// Circular shift
	temp := make([]float64, e.dimensionality)
	for i := 0; i < e.dimensionality; i++ {
		temp[(i+shift)%e.dimensionality] = adjusted.Values[i]
	}
	
	copy(adjusted.Values, temp)
	
	return adjusted
}

func (e *HDComputingEngine) fragmentVector(v *HDVector, chunkSize int) [][]float64 {
	chunks := [][]float64{}
	
	for i := 0; i < e.dimensionality; i += chunkSize {
		end := i + chunkSize
		if end > e.dimensionality {
			end = e.dimensionality
		}
		
		chunk := make([]float64, end-i)
		copy(chunk, v.Values[i:end])
		chunks = append(chunks, chunk)
	}
	
	return chunks
}

func (e *HDComputingEngine) selectAdversarialChunks(chunks [][]float64, iteration int) [][]float64 {
	// Select subset of chunks adversarially
	selected := [][]float64{}
	
	// Use iteration to deterministically select chunks
	rand.Seed(int64(iteration))
	
	// Select 60-80% of chunks
	selectRatio := 0.6 + rand.Float64()*0.2
	numSelect := int(float64(len(chunks)) * selectRatio)
	
	indices := rand.Perm(len(chunks))[:numSelect]
	
	for _, idx := range indices {
		selected = append(selected, chunks[idx])
	}
	
	return selected
}

func (e *HDComputingEngine) holographicReconstruct(chunks [][]float64, targetDim int) *HDVector {
	reconstructed := &HDVector{
		Values:    make([]float64, targetDim),
		Dimension: targetDim,
		Properties: make(map[string]interface{}),
	}
	
	// Holographic reconstruction - each chunk influences entire vector
	for _, chunk := range chunks {
		// Expand chunk to full dimension
		expanded := e.expandChunk(chunk, targetDim)
		
		// Add with holographic interference
		for i := 0; i < targetDim; i++ {
			reconstructed.Values[i] += expanded[i] / float64(len(chunks))
		}
	}
	
	return e.normalizeVector(reconstructed)
}

func (e *HDComputingEngine) expandChunk(chunk []float64, targetDim int) []float64 {
	expanded := make([]float64, targetDim)
	chunkLen := len(chunk)
	
	// Repeat pattern with phase shifts
	for i := 0; i < targetDim; i++ {
		sourceIdx := i % chunkLen
		phase := float64(i) / float64(targetDim) * 2.0 * math.Pi
		expanded[i] = chunk[sourceIdx] * math.Cos(phase)
	}
	
	return expanded
}

func (e *HDComputingEngine) generateHolographicNoise(reference *HDVector, intensity float64) *HDVector {
	// Generate noise that preserves holographic properties
	noise := e.createZeroVector()
	
	// Create interference pattern
	for i := 0; i < e.dimensionality; i++ {
		// Multiple frequency components
		freq1 := float64(i) * 2.0 * math.Pi / float64(e.dimensionality)
		freq2 := float64(i) * 4.0 * math.Pi / float64(e.dimensionality)
		
		noise.Values[i] = intensity * (math.Sin(freq1) + 0.5*math.Sin(freq2))
	}
	
	return noise
}

func (e *HDComputingEngine) measureDistortion(v1, v2 *HDVector) float64 {
	// Measure semantic distortion between vectors
	
	// Basic distance metric
	distance := 0.0
	for i := 0; i < e.dimensionality; i++ {
		diff := v1.Values[i] - v2.Values[i]
		distance += diff * diff
	}
	distance = math.Sqrt(distance / float64(e.dimensionality))
	
	// Structural distortion
	structural := e.measureStructuralDifference(v1, v2)
	
	// Combined distortion measure
	return (distance + structural) / 2.0
}

func (e *HDComputingEngine) evaluateAttackVector(v *HDVector, targets []*HDVector) float64 {
	// Evaluate how effective the attack vector is
	
	maxSim := 0.0
	for _, target := range targets {
		sim := e.cosineSimilarity(v, target)
		if sim > maxSim {
			maxSim = sim
		}
	}
	
	// Also consider adversarial properties
	if e.hasAdversarialProperties(v) {
		maxSim *= 1.2 // Bonus for adversarial characteristics
	}
	
	return math.Min(maxSim, 1.0)
}

func (e *HDComputingEngine) mutateVector(v *HDVector, rate float64) *HDVector {
	mutated := e.copyVector(v)
	
	for i := 0; i < e.dimensionality; i++ {
		if rand.Float64() < rate {
			// Gaussian mutation
			mutated.Values[i] += rand.NormFloat64() * 0.1
		}
	}
	
	return e.normalizeVector(mutated)
}