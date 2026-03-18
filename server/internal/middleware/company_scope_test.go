package middleware

import (
	"context"
	"testing"
)

func TestCompanyFromContext_Empty(t *testing.T) {
	id := CompanyFromContext(context.Background())
	if id != "" {
		t.Errorf("CompanyFromContext empty: expected empty, got %q", id)
	}
}

func TestCompanyFromContext_Set(t *testing.T) {
	ctx := context.WithValue(context.Background(), companyKey, "company-abc")
	id := CompanyFromContext(ctx)
	if id != "company-abc" {
		t.Errorf("CompanyFromContext: expected %q, got %q", "company-abc", id)
	}
}
