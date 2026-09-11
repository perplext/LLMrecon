package testutil

import (
	"context"
	"strings"
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

	// CommandChainErr, when set, is returned from RunCommandChain (SkipProviderError).
	CommandChainErr error
	// ExtraCommandBlock adds tokens to the per-command blocklist so a test can
	// model a hardened agent that rejects one step and breaks the chain.
	ExtraCommandBlock []string
	// ForceNoDangerousComposition suppresses the composition finding, modeling a
	// sandbox that isolates OS state between commands (the secure path).
	ForceNoDangerousComposition bool

	// Recorded effects for assertions.
	Writes      []ResolvedWrite
	ExecutedMCP []string
	RanCommands []string
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

// ---------------------------------------------------------------------------
// MockCodingAgent — CommandChainProvider (v0.13.0, MOSAIC)
// ---------------------------------------------------------------------------

// Compile-time check: MockCodingAgent also satisfies CommandChainProvider.
var _ common.CommandChainProvider = (*MockCodingAgent)(nil)

// commandBlocklist are tokens a per-command filter blocks in isolation. A
// realistic coding agent rejects overtly-destructive commands but waves
// through the individually-benign steps MOSAIC composes.
var commandBlocklist = []string{"rm -rf /", "curl ", "wget ", ":(){", "mkfs", "dd if="}

// producerConsumerExec are tokens marking a command that executes/sources
// shell content — the "consumer" end of a MOSAIC producer→consumer flow.
var producerConsumerExec = []string{"sh -c", "| sh", "| bash", "bash -c", "eval ", "source ", ". ./"}

// RunCommandChain evaluates each command against a per-command blocklist and
// detects a dangerous producer→consumer composition (a write via ">" followed
// by a command that executes shell content). Knobs let tests force outcomes:
//   - CommandChainErr → provider error.
//   - ExtraCommandBlock adds tokens to the per-command blocklist (models a
//     hardened agent that breaks the chain).
//   - ForceNoDangerousComposition suppresses the composition finding (models a
//     sandbox that isolates OS state between commands).
func (m *MockCodingAgent) RunCommandChain(_ context.Context, commands []string) (common.CommandChainOutcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CommandChainErr != nil {
		return common.CommandChainOutcome{}, m.CommandChainErr
	}

	block := append([]string(nil), commandBlocklist...)
	block = append(block, m.ExtraCommandBlock...)

	allowed := make([]bool, len(commands))
	for i, c := range commands {
		ok := true
		for _, bad := range block {
			if bad != "" && strings.Contains(c, bad) {
				ok = false
				break
			}
		}
		allowed[i] = ok
	}
	m.RanCommands = append(m.RanCommands, commands...)

	// Producer→consumer detection: a write ("... > file") followed by a later
	// command that executes/sources shell content.
	dangerous := false
	detail := "no dangerous cross-command composition detected"
	if !m.ForceNoDangerousComposition {
		producerSeen := false
		for _, c := range commands {
			if strings.Contains(c, ">") {
				producerSeen = true
			}
			if producerSeen {
				for _, ex := range producerConsumerExec {
					if strings.Contains(c, ex) {
						dangerous = true
						detail = "producer write followed by consumer exec of shared state"
						break
					}
				}
			}
			if dangerous {
				break
			}
		}
	}

	return common.CommandChainOutcome{
		PerCommandAllowed:    allowed,
		DangerousComposition: dangerous,
		Detail:               detail,
	}, nil
}
