package services

import (
	"context"
	"encoding/json"
	"testing"

	"soksak/internal/domain"
)

// TestResolveCommunicationLanguage verifies the priority chain:
// agent runtime_config → company → instance settings → "en" fallback.
func TestResolveCommunicationLanguage_AgentOverrideTakesPriority(t *testing.T) {
	svc := &AgentService{}
	lang := "ko"
	agent := &domain.Agent{
		RuntimeConfig: mustMarshal(t, map[string]string{"communication_language": "ja"}),
	}
	company := &domain.Company{
		CommunicationLanguage: &lang,
	}
	// Agent "ja" wins over company "ko".
	got, err := svc.ResolveCommunicationLanguage(context.Background(), agent, company, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ja" {
		t.Errorf("got %q, want %q (agent runtime_config should take priority)", got, "ja")
	}
}

func TestResolveCommunicationLanguage_CompanyFallback(t *testing.T) {
	svc := &AgentService{}
	lang := "ko"
	agent := &domain.Agent{
		RuntimeConfig: json.RawMessage("{}"),
	}
	company := &domain.Company{
		CommunicationLanguage: &lang,
	}
	got, err := svc.ResolveCommunicationLanguage(context.Background(), agent, company, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ko" {
		t.Errorf("got %q, want %q (company should be used when agent has no override)", got, "ko")
	}
}

func TestResolveCommunicationLanguage_DefaultFallback(t *testing.T) {
	svc := &AgentService{}
	agent := &domain.Agent{
		RuntimeConfig: json.RawMessage("{}"),
	}
	company := &domain.Company{
		CommunicationLanguage: nil,
	}
	got, err := svc.ResolveCommunicationLanguage(context.Background(), agent, company, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "en" {
		t.Errorf("got %q, want %q (should fall back to 'en')", got, "en")
	}
}

func TestResolveCommunicationLanguage_NilAgentRuntimeConfig(t *testing.T) {
	svc := &AgentService{}
	agent := &domain.Agent{
		RuntimeConfig: nil,
	}
	lang := "ja"
	company := &domain.Company{
		CommunicationLanguage: &lang,
	}
	got, err := svc.ResolveCommunicationLanguage(context.Background(), agent, company, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ja" {
		t.Errorf("got %q, want %q", got, "ja")
	}
}

func mustMarshal(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustMarshal: %v", err)
	}
	return json.RawMessage(b)
}
