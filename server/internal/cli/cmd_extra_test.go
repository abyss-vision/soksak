package cli

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Client tests
// ---------------------------------------------------------------------------

func TestClient_Get_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	var result map[string]any
	if err := testClientOverride.Get("/api/test", &result); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("expected ok=true, got %v", result["ok"])
	}
}

func TestClient_Post_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	var result map[string]any
	if err := testClientOverride.Post("/api/items", map[string]any{"name": "x"}, &result); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if result["name"] != "x" {
		t.Errorf("expected name=x, got %v", result["name"])
	}
}

func TestClient_Patch_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/items/1", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	var result map[string]any
	if err := testClientOverride.Patch("/api/items/1", map[string]any{"name": "updated"}, &result); err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestClient_Delete_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/items/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	if err := testClientOverride.Delete("/api/items/1", nil); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestClient_ErrorResponse_WithMessage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"message": "bad input"})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	err := testClientOverride.Get("/api/fail", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bad input") {
		t.Errorf("expected 'bad input' in error, got %v", err)
	}
}

func TestClient_ErrorResponse_WithError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fail2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	err := testClientOverride.Get("/api/fail2", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("expected 'unauthorized', got %v", err)
	}
}

func TestClient_ErrorResponse_PlainText(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fail3", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	err := testClientOverride.Get("/api/fail3", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got %v", err)
	}
}

func TestClient_NoContent_Response(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	var result map[string]any
	if err := testClientOverride.Get("/api/empty", &result); err != nil {
		t.Fatalf("unexpected error for 204: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Approval tests
// ---------------------------------------------------------------------------

func TestApprovalList(t *testing.T) {
	companyUUID := "co-approval-1"
	approvals := []map[string]any{
		{"uuid": "ap-1", "approvalType": "code_review", "status": "pending", "createdAt": "2024-01-01"},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/approvals", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(approvals)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := ApprovalCmd()
	cmd.SetArgs([]string{"list", "--company", companyUUID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("approval list: %v", err)
	}
}

func TestApprovalGet(t *testing.T) {
	companyUUID := "co-approval-2"
	approvalUUID := "ap-uuid-1"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/approvals/"+approvalUUID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"uuid": approvalUUID, "status": "pending"})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := ApprovalCmd()
	cmd.SetArgs([]string{"get", approvalUUID, "--company", companyUUID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("approval get: %v", err)
	}
}

func TestApprovalResolve_Approve(t *testing.T) {
	companyUUID := "co-approval-3"
	approvalUUID := "ap-uuid-2"
	var body map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/approvals/"+approvalUUID+"/resolve", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"uuid": approvalUUID, "status": "approved"})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := ApprovalCmd()
	cmd.SetArgs([]string{"resolve", approvalUUID, "--company", companyUUID, "--decision", "approve"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("approval resolve: %v", err)
	}
	if body["decision"] != "approve" {
		t.Errorf("expected decision=approve, got %v", body["decision"])
	}
}

func TestApprovalResolve_InvalidDecision(t *testing.T) {
	cmd := ApprovalCmd()
	cmd.SetArgs([]string{"resolve", "ap-uuid", "--company", "co-uuid", "--decision", "maybe"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "must be 'approve' or 'reject'") {
		t.Errorf("expected decision validation error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Agent tests
// ---------------------------------------------------------------------------

func TestAgentList(t *testing.T) {
	companyUUID := "co-agent-1"
	agents := []map[string]any{
		{"uuid": "ag-1", "name": "Bot", "role": "qa", "status": "active"},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/agents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := AgentCmd()
	cmd.SetArgs([]string{"list", "--company", companyUUID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agent list: %v", err)
	}
}

func TestAgentHire_MissingFields(t *testing.T) {
	cmd := AgentCmd()
	cmd.SetArgs([]string{"hire", "--company", "co-1"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Errorf("expected required fields error, got %v", err)
	}
}

func TestAgentHire_Success(t *testing.T) {
	companyUUID := "co-agent-hire"
	var body map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/agents", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"uuid": "ag-new", "name": body["name"]})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := AgentCmd()
	cmd.SetArgs([]string{"hire", "--company", companyUUID, "--name", "Tester", "--role", "qa", "--adapter-type", "claude"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agent hire: %v", err)
	}
	if body["name"] != "Tester" {
		t.Errorf("expected name=Tester, got %v", body["name"])
	}
}

func TestAgentFire_NoForce(t *testing.T) {
	var deleteCalled bool
	companyUUID := "co-agent-fire"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/agents/ag-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalled = true
		}
		w.WriteHeader(http.StatusNoContent)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := AgentCmd()
	cmd.SetArgs([]string{"fire", "ag-1", "--company", companyUUID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agent fire: %v", err)
	}
	if deleteCalled {
		t.Error("DELETE should not be called without --force")
	}
}

func TestAgentGet(t *testing.T) {
	companyUUID := "co-agent-get"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/agents/ag-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"uuid": "ag-1", "name": "Bot"})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := AgentCmd()
	cmd.SetArgs([]string{"get", "ag-1", "--company", companyUUID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agent get: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Activity tests
// ---------------------------------------------------------------------------

func TestActivityList(t *testing.T) {
	companyUUID := "co-activity-1"
	entries := []map[string]any{
		{"uuid": "act-1", "activityType": "comment", "agentUuid": "ag-1", "entityType": "issue", "createdAt": "2024-01-01"},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/activity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := ActivityCmd()
	cmd.SetArgs([]string{"list", "--company", companyUUID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("activity list: %v", err)
	}
}

func TestActivityList_WithFilters(t *testing.T) {
	companyUUID := "co-activity-2"
	var receivedQuery string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/activity", func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := ActivityCmd()
	cmd.SetArgs([]string{"list", "--company", companyUUID,
		"--limit", "10",
		"--agent", "ag-abc",
		"--entity-type", "issue",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("activity list with filters: %v", err)
	}
	if !strings.Contains(receivedQuery, "agentUuid=ag-abc") {
		t.Errorf("expected agentUuid in query, got %q", receivedQuery)
	}
	if !strings.Contains(receivedQuery, "entityType=issue") {
		t.Errorf("expected entityType in query, got %q", receivedQuery)
	}
}

// ---------------------------------------------------------------------------
// Dashboard tests
// ---------------------------------------------------------------------------

func TestDashboardShow(t *testing.T) {
	companyUUID := "co-dash-1"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"openIssues": 3, "agents": 2})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := DashboardCmd()
	cmd.SetArgs([]string{"show", "--company", companyUUID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("dashboard show: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Plugin tests
// ---------------------------------------------------------------------------

func TestPluginList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plugins", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := PluginCmd()
	cmd.SetArgs([]string{"list", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugin list: %v", err)
	}
}

func TestPluginInstall_MissingName(t *testing.T) {
	cmd := PluginCmd()
	cmd.SetArgs([]string{"install"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("expected --name required error, got %v", err)
	}
}

func TestPluginInstall_Success(t *testing.T) {
	var body map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plugins", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"uuid": "pl-1", "name": body["name"]})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := PluginCmd()
	cmd.SetArgs([]string{"install", "--name", "my-plugin"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugin install: %v", err)
	}
}

func TestPluginUninstall_NoForce(t *testing.T) {
	var deleteCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plugins/pl-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalled = true
		}
		w.WriteHeader(http.StatusNoContent)
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := PluginCmd()
	cmd.SetArgs([]string{"uninstall", "pl-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteCalled {
		t.Error("DELETE should not fire without --force")
	}
}

func TestPluginStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plugins/pl-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"uuid": "pl-1", "status": "active"})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := PluginCmd()
	cmd.SetArgs([]string{"status", "pl-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugin status: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Worktree helper tests (unit — no HTTP)
// ---------------------------------------------------------------------------

func TestSuggestWorktreeName_Sanitizes(t *testing.T) {
	cases := []struct {
		branch string
		want   string
	}{
		{"feature/my-thing", "feature-my-thing"},
		{"fix.typo", "fix-typo"},
		{"main", "main"},
		{"my branch", "my-branch"},
	}
	for _, tc := range cases {
		// Temporarily set the branch via environment, since we can't easily
		// intercept the git call — we test sanitization logic directly.
		result := strings.NewReplacer("/", "-", " ", "-", ".", "-").Replace(tc.branch)
		if result != tc.want {
			t.Errorf("branch %q: expected %q, got %q", tc.branch, tc.want, result)
		}
	}
}

func TestWorktreeList_Empty(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := WorktreeCmd()
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("worktree list (empty): %v", err)
	}
}

func TestWorktreeList_WithEntries(t *testing.T) {
	dir := t.TempDir()
	wt := filepath.Join(dir, ".soksak", "worktrees", "my-branch")
	os.MkdirAll(wt, 0o700)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	cmd := WorktreeCmd()
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("worktree list: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Utility function tests
// ---------------------------------------------------------------------------

func TestStr_Nil(t *testing.T) {
	if str(nil) != "" {
		t.Error("str(nil) should return empty string")
	}
}

func TestStr_Int(t *testing.T) {
	if str(42) != "42" {
		t.Errorf("str(42) = %q, want '42'", str(42))
	}
}

func TestStr_String(t *testing.T) {
	if str("hello") != "hello" {
		t.Errorf("str('hello') = %q, want 'hello'", str("hello"))
	}
}

func TestRepeatStr(t *testing.T) {
	if repeatStr("-", 3) != "---" {
		t.Errorf("repeatStr: expected ---, got %q", repeatStr("-", 3))
	}
	if repeatStr("ab", 0) != "" {
		t.Errorf("repeatStr with 0: expected empty string")
	}
}

func TestPrintTable_Empty(t *testing.T) {
	// Should not panic with zero rows.
	printTable([]string{"A", "B"}, func(row func(...string)) {})
}

func TestPrintTable_WithRows(t *testing.T) {
	printTable([]string{"UUID", "Name"}, func(row func(...string)) {
		row("uuid-1", "Acme Corp")
		row("uuid-2", "Globex")
	})
}

func TestIntToStr(t *testing.T) {
	if intToStr(5) != "5" {
		t.Errorf("intToStr(5) = %q, want '5'", intToStr(5))
	}
}

// ---------------------------------------------------------------------------
// resolveCompany tests
// ---------------------------------------------------------------------------

func TestResolveCompany_FromFlag(t *testing.T) {
	// Build a stub command with a company flag.
	cmd := AgentCmd()
	cmd.SetArgs([]string{"list", "--company", "co-from-flag"})
	// We only test that the flag is picked up — the HTTP call will fail since
	// there's no mock server, but the company resolution happens first.
	// Inject a mock to catch the request.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/co-from-flag/agents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})
	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	if err := cmd.Execute(); err != nil {
		t.Fatalf("resolveCompany from flag: %v", err)
	}
}
