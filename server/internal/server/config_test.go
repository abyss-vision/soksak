package server

import "testing"

func TestLoadConfig_Defaults(t *testing.T) {
	// Ensure env vars don't interfere.
	t.Setenv("PORT", "")
	t.Setenv("HOST", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SERVE_UI", "")
	t.Setenv("DEPLOYMENT_MODE", "")
	t.Setenv("SOKSAK_AGENT_JWT_SECRET", "")
	t.Setenv("STORAGE_BASE_DIR", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: unexpected error: %v", err)
	}
	if cfg.Port != "3200" {
		t.Errorf("Port = %q, want %q", cfg.Port, "3200")
	}
	if cfg.Host != "" {
		t.Errorf("Host = %q, want empty", cfg.Host)
	}
	if cfg.ServeUI != false {
		t.Errorf("ServeUI = %v, want false", cfg.ServeUI)
	}
	if cfg.DeploymentMode != "dev" {
		t.Errorf("DeploymentMode = %q, want %q", cfg.DeploymentMode, "dev")
	}
	if cfg.StorageBaseDir != "./data/storage" {
		t.Errorf("StorageBaseDir = %q, want %q", cfg.StorageBaseDir, "./data/storage")
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("HOST", "0.0.0.0")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("SERVE_UI", "true")
	t.Setenv("DEPLOYMENT_MODE", "production")
	t.Setenv("SOKSAK_AGENT_JWT_SECRET", "mysecret")
	t.Setenv("STORAGE_BASE_DIR", "/tmp/storage")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://localhost/test")
	}
	if !cfg.ServeUI {
		t.Error("ServeUI: expected true")
	}
	if cfg.DeploymentMode != "production" {
		t.Errorf("DeploymentMode = %q, want %q", cfg.DeploymentMode, "production")
	}
	if cfg.JWTSecret != "mysecret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "mysecret")
	}
	if cfg.StorageBaseDir != "/tmp/storage" {
		t.Errorf("StorageBaseDir = %q, want %q", cfg.StorageBaseDir, "/tmp/storage")
	}
}

func TestConfig_Addr(t *testing.T) {
	cfg := &Config{Host: "0.0.0.0", Port: "8080"}
	want := "0.0.0.0:8080"
	if got := cfg.Addr(); got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestConfig_Addr_EmptyHost(t *testing.T) {
	cfg := &Config{Host: "", Port: "3200"}
	want := ":3200"
	if got := cfg.Addr(); got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	t.Setenv("TEST_KEY_XYZ", "")
	got := getEnv("TEST_KEY_XYZ", "fallback")
	if got != "fallback" {
		t.Errorf("getEnv fallback = %q, want %q", got, "fallback")
	}
}

func TestGetEnv_Set(t *testing.T) {
	t.Setenv("TEST_KEY_XYZ", "value")
	got := getEnv("TEST_KEY_XYZ", "fallback")
	if got != "value" {
		t.Errorf("getEnv set = %q, want %q", got, "value")
	}
}

func TestGetEnvBool(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	if !getEnvBool("TEST_BOOL", false) {
		t.Error("getEnvBool true: expected true")
	}

	t.Setenv("TEST_BOOL", "false")
	if getEnvBool("TEST_BOOL", true) {
		t.Error("getEnvBool false: expected false")
	}

	t.Setenv("TEST_BOOL", "")
	if !getEnvBool("TEST_BOOL", true) {
		t.Error("getEnvBool empty: expected fallback true")
	}

	t.Setenv("TEST_BOOL", "invalid")
	if getEnvBool("TEST_BOOL", true) != true {
		t.Error("getEnvBool invalid: expected fallback true")
	}
}
