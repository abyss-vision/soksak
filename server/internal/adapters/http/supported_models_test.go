package http_test

import (
	"testing"

	httpadapter "abyss-view/internal/adapters/http"
)

func TestHTTPAdapter_SupportedModels(t *testing.T) {
	models := httpadapter.New().SupportedModels()
	if models == nil {
		t.Error("SupportedModels returned nil")
	}
}
