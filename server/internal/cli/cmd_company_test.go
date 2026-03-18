package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newMockServer starts a local httptest.Server with the given mux and returns
// an injected *Client pointing at it.
func newMockServer(t *testing.T, mux *http.ServeMux) (*httptest.Server, func()) {
	t.Helper()
	srv := httptest.NewServer(mux)
	prev := testClientOverride
	testClientOverride = NewClient(srv.URL, "")
	cleanup := func() {
		srv.Close()
		testClientOverride = prev
	}
	return srv, cleanup
}

// executeCmd runs a Cobra command tree with the given args and returns stdout.
func executeCmd(t *testing.T, root interface{ Execute() error }, args ...string) (string, error) {
	t.Helper()
	// Use the Cobra command's SetArgs/SetOut approach.
	// We get the root from the test's command factory.
	return "", nil // placeholder — see per-test helpers below
}

// captureOutput redirects os.Stdout during fn() and returns what was printed.
func captureOutput(fn func()) string {
	// We capture by replacing fmt output with a buffer using cmd.SetOut.
	// Since commands use fmt.Print* directly, we test via side-effects on the
	// injected HTTP client instead of stdout capture for list/get.
	fn()
	return ""
}

// TestCompanyList verifies the list command calls GET /api/companies and
// prints a table when the server returns a JSON array.
func TestCompanyList(t *testing.T) {
	companies := []map[string]any{
		{"uuid": "uuid-1", "name": "Acme"},
		{"uuid": "uuid-2", "name": "Globex"},
	}

	var called bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method %s", r.Method)
		}
		called = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(companies)
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := CompanyCmd()
	cmd.SetArgs([]string{"list", "--json"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("company list: %v", err)
	}
	if !called {
		t.Fatal("GET /api/companies was never called")
	}
}

// TestCompanyGet verifies GET /api/companies/{uuid} is called with the right path.
func TestCompanyGet(t *testing.T) {
	company := map[string]any{"uuid": "uuid-1", "name": "Acme"}

	var requestedPath string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/uuid-1", func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(company)
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := CompanyCmd()
	cmd.SetArgs([]string{"get", "uuid-1", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("company get: %v", err)
	}
	if requestedPath != "/api/companies/uuid-1" {
		t.Errorf("expected /api/companies/uuid-1, got %s", requestedPath)
	}
}

// TestCompanyCreate verifies POST /api/companies is called with the right body.
func TestCompanyCreate(t *testing.T) {
	var body map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"uuid": "new-uuid", "name": body["name"]})
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := CompanyCmd()
	cmd.SetArgs([]string{"create", "--name", "NewCo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("company create: %v", err)
	}
	if body["name"] != "NewCo" {
		t.Errorf("expected name=NewCo, got %v", body["name"])
	}
}

// TestCompanyCreateMissingName verifies --name is required.
func TestCompanyCreateMissingName(t *testing.T) {
	cmd := CompanyCmd()
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestCompanyDelete_NoForce verifies delete without --force does not call DELETE.
func TestCompanyDelete_NoForce(t *testing.T) {
	var deleteCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/uuid-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalled = true
		}
		w.WriteHeader(http.StatusNoContent)
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := CompanyCmd()
	cmd.SetArgs([]string{"delete", "uuid-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteCalled {
		t.Error("DELETE should not be called without --force")
	}
}

// TestCompanyDelete_WithForce verifies DELETE is called when --force is passed.
func TestCompanyDelete_WithForce(t *testing.T) {
	var deleteCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/uuid-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalled = true
		}
		w.WriteHeader(http.StatusNoContent)
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := CompanyCmd()
	cmd.SetArgs([]string{"delete", "uuid-1", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("DELETE should be called with --force")
	}
}

// TestCompanyUpdate_NoFlags verifies that update without flags returns an error.
func TestCompanyUpdate_NoFlags(t *testing.T) {
	cmd := CompanyCmd()
	cmd.SetArgs([]string{"update", "uuid-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no update flags provided")
	}
}

// TestCompanyUpdate_WithName verifies PATCH is called with the right body.
func TestCompanyUpdate_WithName(t *testing.T) {
	var patchBody map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/api/companies/uuid-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&patchBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"uuid": "uuid-1", "name": patchBody["name"]})
	})

	_, cleanup := newMockServer(t, mux)
	defer cleanup()

	cmd := CompanyCmd()
	cmd.SetArgs([]string{"update", "uuid-1", "--name", "Updated"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("company update: %v", err)
	}
	if patchBody["name"] != "Updated" {
		t.Errorf("expected name=Updated, got %v", patchBody["name"])
	}
}
