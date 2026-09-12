// Ghost Vectors — reconstruction of soft-deleted embeddings in HNSW vector
// databases (arXiv 2606.18497, "Ghost Vectors: Soft-Deleted Embeddings Remain
// Reconstructible in HNSW Vector Databases", Jun 2026).
//
// HNSW-based vector stores implement deletion as a soft-delete: the record is
// tombstoned but its raw embedding stays in the index files until a compaction
// that many deployments never run. An attacker with storage-layer access reads
// the raw index, extracts the "deleted" vectors, and runs Vec2Text-style
// inversion to reconstruct the source text — bypassing the store's API-level
// delete and any right-to-be-forgotten guarantee built on it. The paper reports
// up to 99–100% recovery in some domains.
//
// The vulnerability is deletion DURABILITY, not the inversion model. This
// module tests exactly that: insert a canary → delete it through the normal
// API → read the raw index → success only if the soft-deleted canary's source
// text is still recoverable. It targets a storage-layer surface rather than an
// LLM's text output, so it requires common.VectorStoreProbe; against a plain
// provider it emits OutcomeSkipped + SkipMissingCapability, and the
// testutil.MockVectorStore double exercises it end-to-end.
package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
)

func init() {
	attacks.DefaultRegistry.Register(&GhostVectorsModule{})
}

// GhostVectorsModule implements the AttackModule interface for Ghost Vectors.
type GhostVectorsModule struct{}

// Name returns the registered technique name.
func (m *GhostVectorsModule) Name() string { return "ghost_vectors" }

// Category returns CategoryRAG.
func (m *GhostVectorsModule) Category() common.AttackCategory { return common.CategoryRAG }

// Description summarizes the technique.
func (m *GhostVectorsModule) Description() string {
	return "Ghost Vectors — reconstruct soft-deleted embeddings from an HNSW vector store's raw index, bypassing API-level deletion (arXiv 2606.18497)."
}

// Techniques returns the OWASP and metadata bundle.
func (m *GhostVectorsModule) Techniques() []common.TechniqueInfo {
	return []common.TechniqueInfo{{
		ID:                     "ghost_vectors",
		Name:                   "Ghost Vectors",
		Description:            m.Description(),
		Category:               string(common.CategoryRAG),
		Risk:                   "high",
		OWASPLLMCategories:     []string{"LLM02:2025", "LLM08:2025"},
		OWASPAgenticCategories: []string{"ASI04"},
	}}
}

// Execute runs the deletion-durability probe against a vector-store target.
//
//  1. Gate — config.Metadata["i_understand_risks"] must equal "true".
//  2. Capability — provider must implement common.VectorStoreProbe.
//  3. Insert a canary record, delete it through the normal API, read the raw
//     index.
//  4. Success when the soft-deleted canary is still present in the raw index
//     with its source text recoverable; refused when the store compacted the
//     record away (nothing recoverable). A raw-index read error is a
//     SkipProviderError, never a fabricated refusal.
func (m *GhostVectorsModule) Execute(
	ctx context.Context,
	provider common.Provider,
	config common.AttackConfig,
) (*common.AttackResult, error) {
	start := time.Now()
	skipped := func(reason common.SkipReason, detail string) *common.AttackResult {
		r := common.NewAttackResult(m.Name(), common.OutcomeSkipped)
		r.WithSkip(reason, detail)
		r.Duration = time.Since(start)
		return r
	}

	// 1. Safety gate.
	if config.Metadata["i_understand_risks"] != "true" {
		r := skipped(common.SkipGateBlocked, "ghost_vectors requires i_understand_risks=true")
		r.FailureReasons = append(r.FailureReasons,
			"ghost_vectors writes and deletes real records in the target store; operator must opt in")
		return r, nil
	}

	// 2. Capability probe.
	store, ok := provider.(common.VectorStoreProbe)
	if !ok {
		return skipped(common.SkipMissingCapability, "common.VectorStoreProbe"), nil
	}

	// 3. Insert a canary, then delete it through the normal API.
	canaryID := "ghostvec-" + common.GenerateAttackID()
	marker := "GHOSTVEC-CANARY-" + canaryID
	recordBody := config.Payload
	if recordBody == "" {
		recordBody = "synthetic record body flagged for right-to-be-forgotten deletion"
	}
	canaryText := fmt.Sprintf("%s :: %s", marker, recordBody)

	if err := store.InsertVector(ctx, canaryID, canaryText); err != nil {
		return skipped(common.SkipProviderError, fmt.Sprintf("InsertVector: %v", err)), nil
	}
	if err := store.DeleteVector(ctx, canaryID); err != nil {
		return skipped(common.SkipProviderError, fmt.Sprintf("DeleteVector: %v", err)), nil
	}

	// 4. Read the raw index and look for the soft-deleted canary.
	raw, err := store.ReadRawIndex(ctx)
	if err != nil {
		return skipped(common.SkipProviderError, fmt.Sprintf("ReadRawIndex: %v", err)), nil
	}

	var recovered *common.RawVectorRecord
	for i := range raw {
		if raw[i].ID == canaryID {
			recovered = &raw[i]
			break
		}
	}

	// Deletion-durability success requires the canary to be BOTH recoverable
	// (marker present in the raw index) AND actually tombstoned. A live,
	// undeleted record (Deleted=false) means the delete never took effect — a
	// broken-delete finding, not the soft-delete-durability the module claims.
	var result *common.AttackResult
	recoveredMarker := recovered != nil && strings.Contains(recovered.Text, marker)
	switch {
	case recoveredMarker && recovered.Deleted:
		result = common.NewAttackResult(m.Name(), common.OutcomeSuccess)
		result.Confidence = 0.95
		result.Response = recovered.Text
		result.SuccessFactors = append(result.SuccessFactors,
			fmt.Sprintf("tombstoned canary %q still reconstructible from the raw index after API delete — soft-deleted data recovery", canaryID))
	case recoveredMarker && !recovered.Deleted:
		// Delete API returned success but left the record live and untombstoned.
		result = common.NewAttackResult(m.Name(), common.OutcomeRefused)
		result.FailureReasons = append(result.FailureReasons,
			"canary is present but not tombstoned — DeleteVector did not take effect (not a soft-delete durability finding)")
	default:
		result = common.NewAttackResult(m.Name(), common.OutcomeRefused)
		result.FailureReasons = append(result.FailureReasons,
			"deleted canary was not recoverable from the raw index (store compacts on delete)")
	}

	result.Payload = fmt.Sprintf("insert+delete canary %s", canaryID)
	result.AttemptCount = 1
	if result.Metadata == nil {
		result.Metadata = map[string]interface{}{}
	}
	result.Metadata["canary_id"] = canaryID
	result.Metadata["raw_index_size"] = len(raw)
	result.Metadata["recovered"] = recovered != nil
	if recovered != nil {
		result.Metadata["tombstoned"] = recovered.Deleted
	}
	result.Duration = time.Since(start)
	return result, nil
}
