package prompt

import (
	"context"
	"errors"
	"testing"
)

func TestRequestApproval_CallbackApproves(t *testing.T) {
	cfg := DefaultProtectionConfig()
	cfg.ApprovalCallback = func(_ context.Context, _ *ApprovalRequest) (bool, error) {
		return true, nil
	}
	w := NewApprovalWorkflow(cfg)

	approved, _, err := w.RequestApproval(context.Background(), &ProtectionResult{RiskScore: 0.9})
	if err != nil {
		t.Fatalf("RequestApproval: %v", err)
	}
	if !approved {
		t.Fatal("callback returned true, expected approval")
	}
}

func TestRequestApproval_CallbackDenies(t *testing.T) {
	cfg := DefaultProtectionConfig()
	cfg.ApprovalCallback = func(_ context.Context, _ *ApprovalRequest) (bool, error) {
		return false, nil
	}
	w := NewApprovalWorkflow(cfg)

	approved, _, err := w.RequestApproval(context.Background(), &ProtectionResult{RiskScore: 0.9})
	if err != nil {
		t.Fatalf("RequestApproval: %v", err)
	}
	if approved {
		t.Fatal("callback returned false, expected denial")
	}
}

func TestRequestApproval_CallbackError(t *testing.T) {
	cfg := DefaultProtectionConfig()
	cfg.ApprovalCallback = func(_ context.Context, _ *ApprovalRequest) (bool, error) {
		return false, errors.New("boom")
	}
	w := NewApprovalWorkflow(cfg)

	if _, _, err := w.RequestApproval(context.Background(), &ProtectionResult{RiskScore: 0.9}); err == nil {
		t.Fatal("callback error must propagate")
	}
}

func TestRequestApproval_AutoApproveLowRisk(t *testing.T) {
	cfg := DefaultProtectionConfig()
	w := NewApprovalWorkflow(cfg)
	w.workflowConfig.EnableAutoApproval = true // enable the auto-approval path

	approved, res, err := w.RequestApproval(context.Background(), &ProtectionResult{RiskScore: 0.3})
	if err != nil {
		t.Fatalf("RequestApproval: %v", err)
	}
	if !approved {
		t.Fatal("low-risk request should auto-approve")
	}
	if len(res.Detections) == 0 {
		t.Fatal("auto-approval should annotate the result")
	}
}

func TestManualRequestManagement(t *testing.T) {
	w := NewApprovalWorkflow(DefaultProtectionConfig())

	// Seed a pending request directly to exercise the management methods.
	w.pendingRequests["req-1"] = &ApprovalRequest{RequestID: "req-1"}
	w.pendingRequests["req-2"] = &ApprovalRequest{RequestID: "req-2"}

	if got := w.GetPendingRequests(); len(got) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(got))
	}

	if status, ok := w.GetRequestStatus("req-1"); !ok || status != "pending" {
		t.Fatalf("req-1 should be pending, got %q/%v", status, ok)
	}

	if !w.ApproveRequest("req-1") {
		t.Fatal("ApproveRequest on pending should succeed")
	}
	if status, _ := w.GetRequestStatus("req-1"); status != "approved" {
		t.Fatalf("req-1 should be approved, got %q", status)
	}

	if !w.DenyRequest("req-2") {
		t.Fatal("DenyRequest on pending should succeed")
	}
	if status, _ := w.GetRequestStatus("req-2"); status != "denied" {
		t.Fatalf("req-2 should be denied, got %q", status)
	}

	// Acting on an unknown / already-resolved request returns false.
	if w.ApproveRequest("req-1") {
		t.Fatal("ApproveRequest on non-pending must return false")
	}
	if w.DenyRequest("ghost") {
		t.Fatal("DenyRequest on unknown must return false")
	}
	if _, ok := w.GetRequestStatus("ghost"); ok {
		t.Fatal("unknown request must report not found")
	}
}

func TestGetTopDetections(t *testing.T) {
	dets := []*Detection{
		{Description: "a", Confidence: 0.3},
		{Description: "b", Confidence: 0.9},
		{Description: "c", Confidence: 0.6},
	}
	top := getTopDetections(dets, 2)
	if len(top) != 2 {
		t.Fatalf("expected 2, got %d", len(top))
	}
	if top[0].Confidence != 0.9 || top[1].Confidence != 0.6 {
		t.Fatalf("getTopDetections not sorted by confidence desc: %+v", top)
	}
	// n larger than slice returns all.
	if got := getTopDetections(dets, 10); len(got) != 3 {
		t.Fatalf("expected all 3, got %d", len(got))
	}
}

func TestGenerateApprovalReason(t *testing.T) {
	none := generateApprovalReason(&ProtectionResult{RiskScore: 0.9})
	if none == "" {
		t.Fatal("reason should be non-empty even with no detections")
	}
	with := generateApprovalReason(&ProtectionResult{
		RiskScore:  0.9,
		Detections: []*Detection{{Description: "jailbreak", Confidence: 0.95}},
	})
	if with == none {
		t.Fatal("reason should reflect detections when present")
	}
}
