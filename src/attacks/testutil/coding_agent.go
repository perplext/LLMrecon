package testutil

import (
	"context"
	"sync"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// ---------------------------------------------------------------------------
// MockCodingAgent
// ---------------------------------------------------------------------------

// MockCodingAgent is a controllable in-memory implementation of
// common.CodingAgentProvider for exercising the v0.12.0 SymJack and TrustFall
// modules end-to-end. It performs no real filesystem or process operations —
// symlink resolution and MCP auto-execution are simulated and recorded for
// assertion. Refuse-mode knobs let tests drive the OutcomeRefused and
// SkipNoMutationTarget paths, not just success.
type MockCodingAgent struct {
	MockProvider // base common.Provider behavior (Query/GetName/...)

	mu sync.Mutex

	// Symlinks maps a shown destination path to the real path the agent
	// resolves it to. When a FileOperation's ShownDestination matches a key,
	// the write resolves to the mapped target (the SymJack misrepresentation).
	// Absent a mapping, the resolved destination equals the shown destination.
	Symlinks map[string]string

	// NoApprovalStep makes ApproveFileOperation report HasApprovalStep=false
	// (the SkipNoMutationTarget case — agent has nothing to hijack).
	NoApprovalStep bool
	// DenyApproval makes the agent decline the operation (OutcomeRefused).
	DenyApproval bool
	// ApproveErr, when set, is returned from ApproveFileOperation (SkipProviderError).
	ApproveErr error

	// AutoExecuteOnTrust executes the repo's ProjectMCPPaths on a trust accept
	// (the TrustFall default-trust behavior). When false, trust does not
	// auto-execute project MCP (OutcomeRefused).
	AutoExecuteOnTrust bool
	// NoTrustPrompt makes TrustFolder report HasTrustPrompt=false.
	NoTrustPrompt bool
	// TrustErr, when set, is returned from TrustFolder (SkipProviderError).
	TrustErr error

	// Recorded effects for assertions.
	Writes      []ResolvedWrite
	ExecutedMCP []string
}

// ResolvedWrite records a file write the agent performed, capturing the gap
// between what the approval prompt showed and where the bytes actually landed.
type ResolvedWrite struct {
	Source   string
	Shown    string
	Resolved string
}

// Compile-time check: MockCodingAgent satisfies the capability interface.
var _ common.CodingAgentProvider = (*MockCodingAgent)(nil)

// ApproveFileOperation simulates presenting a file operation to the agent's
// approval surface, resolving the destination through the configured symlink
// table and recording any write.
func (m *MockCodingAgent) ApproveFileOperation(_ context.Context, op common.FileOperation) (common.ApprovalOutcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ApproveErr != nil {
		return common.ApprovalOutcome{}, m.ApproveErr
	}
	if m.NoApprovalStep {
		return common.ApprovalOutcome{HasApprovalStep: false}, nil
	}

	resolved := op.ShownDestination
	if target, ok := m.Symlinks[op.ShownDestination]; ok {
		resolved = target
	}
	if m.DenyApproval {
		return common.ApprovalOutcome{HasApprovalStep: true, Approved: false, ResolvedDestination: resolved}, nil
	}
	m.Writes = append(m.Writes, ResolvedWrite{Source: op.SourcePath, Shown: op.ShownDestination, Resolved: resolved})
	return common.ApprovalOutcome{HasApprovalStep: true, Approved: true, ResolvedDestination: resolved, Wrote: true}, nil
}

// TrustFolder simulates a folder-trust decision, auto-executing the repo's
// project MCP paths when AutoExecuteOnTrust is set.
func (m *MockCodingAgent) TrustFolder(_ context.Context, req common.FolderTrustRequest) (common.FolderTrustOutcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.TrustErr != nil {
		return common.FolderTrustOutcome{}, m.TrustErr
	}
	if m.NoTrustPrompt {
		return common.FolderTrustOutcome{HasTrustPrompt: false}, nil
	}
	out := common.FolderTrustOutcome{HasTrustPrompt: true, Trusted: true}
	if m.AutoExecuteOnTrust {
		out.ExecutedPaths = append([]string(nil), req.ProjectMCPPaths...)
		m.ExecutedMCP = append(m.ExecutedMCP, req.ProjectMCPPaths...)
	}
	return out, nil
}

// WriteCount returns the number of recorded writes (for assertions).
func (m *MockCodingAgent) WriteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Writes)
}
