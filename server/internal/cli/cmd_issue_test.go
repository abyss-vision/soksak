package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestIssueTransition_ValidTransition verifies that a valid state machine
// transition results in a PATCH call to the server.
func TestIssueTransition_ValidTransition(t *testing.T) {
	issueUUID := "issue-uuid-1"
	companyUUID := "company-uuid-1"

	var patchCalled bool
	var patchBody map[string]any

	mux := http.NewServeMux()
	// GET to fetch current status
	mux.HandleFunc("/api/companies/"+companyUUID+"/issues/"+issueUUID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"uuid":   issueUUID,
				"status": "open",
			})
		case http.MethodPatch:
			patchCalled = true
			json.NewDecoder(r.Body).Decode(&patchBody)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"uuid": issueUUID, "status": patchBody["status"]})
		}
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := IssueCmd()
	cmd.SetArgs([]string{
		"transition", issueUUID, "in_progress",
		"--company", companyUUID,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("issue transition: %v", err)
	}
	if !patchCalled {
		t.Fatal("PATCH was not called for valid transition")
	}
	if patchBody["status"] != "in_progress" {
		t.Errorf("expected status=in_progress, got %v", patchBody["status"])
	}
}

// TestIssueTransition_InvalidTransition verifies that an illegal state
// transition is rejected with a non-nil error and no PATCH call.
func TestIssueTransition_InvalidTransition(t *testing.T) {
	issueUUID := "issue-uuid-2"
	companyUUID := "company-uuid-1"

	var patchCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/issues/"+issueUUID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			// Issue is in "open" state — "closed" -> "in_progress" is not a valid
			// transition from "open" in the Kanban sense; we test "open" -> "closed"
			// is valid but "closed" -> "in_progress" (from open) is not.
			json.NewEncoder(w).Encode(map[string]any{
				"uuid":   issueUUID,
				"status": "open",
			})
		case http.MethodPatch:
			patchCalled = true
			w.WriteHeader(http.StatusOK)
		}
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := IssueCmd()
	cmd.SetArgs([]string{
		"transition", issueUUID, "blocked", // open -> blocked is NOT allowed
		"--company", companyUUID,
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid transition open -> blocked")
	}
	if !strings.Contains(err.Error(), "invalid transition") {
		t.Errorf("expected 'invalid transition' in error, got: %v", err)
	}
	if patchCalled {
		t.Error("PATCH should not be called for invalid transition")
	}
}

// TestIssueTransition_AllValidPaths exercises every allowed edge in the state machine.
func TestIssueTransition_AllValidPaths(t *testing.T) {
	validCases := []struct {
		from string
		to   string
	}{
		{"open", "in_progress"},
		{"open", "closed"},
		{"in_progress", "open"},
		{"in_progress", "closed"},
		{"in_progress", "blocked"},
		{"blocked", "in_progress"},
		{"blocked", "closed"},
		{"closed", "open"},
	}

	for _, tc := range validCases {
		t.Run(tc.from+"->"+tc.to, func(t *testing.T) {
			issueUUID := "issue-123"
			companyUUID := "company-456"

			mux := http.NewServeMux()
			mux.HandleFunc("/api/companies/"+companyUUID+"/issues/"+issueUUID, func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]any{"uuid": issueUUID, "status": tc.from})
				case http.MethodPatch:
					var body map[string]any
					json.NewDecoder(r.Body).Decode(&body)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]any{"uuid": issueUUID, "status": body["status"]})
				}
			})

			_, cleanup := newMockServer(t, mux)
			defer cleanup()

			cmd := IssueCmd()
			cmd.SetArgs([]string{"transition", issueUUID, tc.to, "--company", companyUUID})
			if err := cmd.Execute(); err != nil {
				t.Errorf("valid transition %s->%s failed: %v", tc.from, tc.to, err)
			}
		})
	}
}

// TestIssueTransition_AllInvalidPaths verifies that every non-allowed edge is rejected.
func TestIssueTransition_AllInvalidPaths(t *testing.T) {
	allStatuses := []string{"open", "in_progress", "blocked", "closed"}
	invalidCases := []struct{ from, to string }{}

	for _, from := range allStatuses {
		allowed := issueTransitions[from]
		allowedSet := map[string]bool{}
		for _, a := range allowed {
			allowedSet[a] = true
		}
		for _, to := range allStatuses {
			if from != to && !allowedSet[to] {
				invalidCases = append(invalidCases, struct{ from, to string }{from, to})
			}
		}
	}

	for _, tc := range invalidCases {
		t.Run(tc.from+"->"+tc.to, func(t *testing.T) {
			issueUUID := "issue-789"
			companyUUID := "company-789"

			mux := http.NewServeMux()
			mux.HandleFunc("/api/companies/"+companyUUID+"/issues/"+issueUUID, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]any{"uuid": issueUUID, "status": tc.from})
				}
			})

			_, cleanup := newMockServer(t, mux)
			defer cleanup()

			cmd := IssueCmd()
			cmd.SetArgs([]string{"transition", issueUUID, tc.to, "--company", companyUUID})
			err := cmd.Execute()
			if err == nil {
				t.Errorf("expected error for invalid transition %s->%s", tc.from, tc.to)
			}
		})
	}
}

// TestIssueCreate_MissingTitle verifies --title is required.
func TestIssueCreate_MissingTitle(t *testing.T) {
	cmd := IssueCmd()
	cmd.SetArgs([]string{"create", "--company", "company-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --title is missing")
	}
}

// TestIssueCreate_Success verifies POST /api/companies/{uuid}/issues is called.
func TestIssueCreate_Success(t *testing.T) {
	companyUUID := "company-create-1"
	var body map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/"+companyUUID+"/issues", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"uuid": "new-issue", "title": body["title"]})
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := IssueCmd()
	cmd.SetArgs([]string{"create", "--title", "Bug fix", "--company", companyUUID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("issue create: %v", err)
	}
	if body["title"] != "Bug fix" {
		t.Errorf("expected title='Bug fix', got %v", body["title"])
	}
}
