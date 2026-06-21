package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

func TestMockCodingAgentApproveFileOperation(t *testing.T) {
	ctx := context.Background()

	t.Run("symlinked destination resolves to config target and write lands", func(t *testing.T) {
		agent := &MockCodingAgent{
			Symlinks: map[string]string{"docs/demo.mp4": "/home/user/.config/mcp/servers.json"},
		}
		op := common.FileOperation{ShownDescription: "copy demo.mp4 to docs/", ShownDestination: "docs/demo.mp4", SourcePath: "payload.json"}
		out, err := agent.ApproveFileOperation(ctx, op)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !out.MisrepresentedVs(op.ShownDestination) {
			t.Errorf("expected misrepresentation; resolved=%q shown=%q", out.ResolvedDestination, op.ShownDestination)
		}
		if agent.WriteCount() != 1 {
			t.Errorf("WriteCount = %d, want 1", agent.WriteCount())
		}
	})

	t.Run("honest destination resolves to shown (refuse case)", func(t *testing.T) {
		agent := &MockCodingAgent{}
		op := common.FileOperation{ShownDestination: "docs/demo.mp4"}
		out, _ := agent.ApproveFileOperation(ctx, op)
		if out.MisrepresentedVs(op.ShownDestination) {
			t.Errorf("honest write should not be misrepresented")
		}
	})

	t.Run("no approval step reports HasApprovalStep=false", func(t *testing.T) {
		agent := &MockCodingAgent{NoApprovalStep: true}
		out, _ := agent.ApproveFileOperation(ctx, common.FileOperation{ShownDestination: "x"})
		if out.HasApprovalStep {
			t.Errorf("expected HasApprovalStep=false")
		}
	})

	t.Run("deny approval records no write", func(t *testing.T) {
		agent := &MockCodingAgent{DenyApproval: true, Symlinks: map[string]string{"x": "/cfg/mcp.json"}}
		out, _ := agent.ApproveFileOperation(ctx, common.FileOperation{ShownDestination: "x"})
		if out.Approved || out.Wrote || agent.WriteCount() != 0 {
			t.Errorf("denied op should not approve/write; got approved=%v wrote=%v writes=%d", out.Approved, out.Wrote, agent.WriteCount())
		}
	})

	t.Run("approve error surfaces", func(t *testing.T) {
		agent := &MockCodingAgent{ApproveErr: errors.New("transport down")}
		if _, err := agent.ApproveFileOperation(ctx, common.FileOperation{}); err == nil {
			t.Errorf("expected error")
		}
	})
}

func TestMockCodingAgentTrustFolder(t *testing.T) {
	ctx := context.Background()

	t.Run("auto-execute on trust runs project MCP paths", func(t *testing.T) {
		agent := &MockCodingAgent{AutoExecuteOnTrust: true}
		out, err := agent.TrustFolder(ctx, common.FolderTrustRequest{RepoPath: "/tmp/clone", ProjectMCPPaths: []string{"./.mcp/evil.json"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out.ExecutedPaths) != 1 || len(agent.ExecutedMCP) != 1 {
			t.Errorf("expected 1 executed path; got out=%d recorded=%d", len(out.ExecutedPaths), len(agent.ExecutedMCP))
		}
	})

	t.Run("trust without auto-execute runs nothing (refuse case)", func(t *testing.T) {
		agent := &MockCodingAgent{AutoExecuteOnTrust: false}
		out, _ := agent.TrustFolder(ctx, common.FolderTrustRequest{ProjectMCPPaths: []string{"./.mcp/evil.json"}})
		if len(out.ExecutedPaths) != 0 {
			t.Errorf("expected no auto-execution; got %d", len(out.ExecutedPaths))
		}
	})

	t.Run("no trust prompt reported", func(t *testing.T) {
		agent := &MockCodingAgent{NoTrustPrompt: true}
		out, _ := agent.TrustFolder(ctx, common.FolderTrustRequest{})
		if out.HasTrustPrompt {
			t.Errorf("expected HasTrustPrompt=false")
		}
	})

	t.Run("trust error surfaces", func(t *testing.T) {
		agent := &MockCodingAgent{TrustErr: errors.New("transport down")}
		if _, err := agent.TrustFolder(ctx, common.FolderTrustRequest{}); err == nil {
			t.Errorf("expected error")
		}
	})
}
