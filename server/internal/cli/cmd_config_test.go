package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// withTempConfigDir overrides the Viper config file to use a temp directory,
// runs fn, then restores Viper state.
func withTempConfigDir(t *testing.T, fn func(cfgFile string)) {
	t.Helper()
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.json")

	// Reset viper for each test to avoid state leaking between tests.
	viper.Reset()
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("json")
	viper.SetDefault("server.url", "http://localhost:3100")
	viper.SetDefault("server.token", "")
	viper.SetDefault("company.default", "")

	fn(cfgFile)

	viper.Reset()
}

// TestWriteConfigValue_PersistsToFile verifies that WriteConfigValue writes
// the key/value to the JSON config file on disk.
func TestWriteConfigValue_PersistsToFile(t *testing.T) {
	withTempConfigDir(t, func(cfgFile string) {
		// Patch the internal configFilePath to use the temp file.
		origHome := os.Getenv("HOME")
		// Redirect HOME so configFilePath() returns our temp path.
		tempHome := filepath.Dir(filepath.Dir(cfgFile)) // parent of ".soksak"
		_ = os.MkdirAll(filepath.Join(tempHome, ".soksak"), 0o700)
		tempCfg := filepath.Join(tempHome, ".soksak", "config.json")
		viper.SetConfigFile(tempCfg)
		t.Cleanup(func() { os.Setenv("HOME", origHome) })
		os.Setenv("HOME", tempHome)

		if err := WriteConfigValue("server.url", "http://myserver:9000"); err != nil {
			t.Fatalf("WriteConfigValue: %v", err)
		}

		data, err := os.ReadFile(tempCfg)
		if err != nil {
			t.Fatalf("read config file: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("parse config JSON: %v", err)
		}
		server, ok := raw["server"].(map[string]any)
		if !ok {
			t.Fatalf("expected server section in config, got %v", raw)
		}
		if server["url"] != "http://myserver:9000" {
			t.Errorf("expected url=http://myserver:9000, got %v", server["url"])
		}
	})
}

// TestWriteConfigValue_NestedKeys verifies multi-level dot-notation keys.
func TestWriteConfigValue_NestedKeys(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, ".soksak", "config.json")
	os.MkdirAll(filepath.Dir(cfgFile), 0o700)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", origHome); viper.Reset() })

	viper.Reset()
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("json")

	if err := WriteConfigValue("company.default", "co-uuid-abc"); err != nil {
		t.Fatalf("WriteConfigValue: %v", err)
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var raw map[string]any
	json.Unmarshal(data, &raw)

	company, ok := raw["company"].(map[string]any)
	if !ok {
		t.Fatalf("expected company section, got %v", raw)
	}
	if company["default"] != "co-uuid-abc" {
		t.Errorf("expected default=co-uuid-abc, got %v", company["default"])
	}
}

// TestWriteConfigValue_PreservesExistingKeys verifies that writing a new key
// does not overwrite unrelated keys already in the file.
func TestWriteConfigValue_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, ".soksak", "config.json")
	os.MkdirAll(filepath.Dir(cfgFile), 0o700)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", origHome); viper.Reset() })

	// Seed file with an existing key.
	initial := `{"server":{"url":"http://existing:3100","token":"tok123"}}`
	os.WriteFile(cfgFile, []byte(initial), 0o600)

	viper.Reset()
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("json")
	viper.ReadInConfig()

	// Write a different key.
	if err := WriteConfigValue("company.default", "co-new"); err != nil {
		t.Fatalf("WriteConfigValue: %v", err)
	}

	data, _ := os.ReadFile(cfgFile)
	var raw map[string]any
	json.Unmarshal(data, &raw)

	server, _ := raw["server"].(map[string]any)
	if server["url"] != "http://existing:3100" {
		t.Errorf("existing server.url was overwritten: %v", server["url"])
	}
	if server["token"] != "tok123" {
		t.Errorf("existing server.token was overwritten: %v", server["token"])
	}
	company, _ := raw["company"].(map[string]any)
	if company["default"] != "co-new" {
		t.Errorf("expected company.default=co-new, got %v", company["default"])
	}
}

// TestSetNestedKey_TopLevel verifies a key with no dot is set at the top level.
func TestSetNestedKey_TopLevel(t *testing.T) {
	m := map[string]any{}
	setNestedKey(m, "foo", "bar")
	if m["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", m["foo"])
	}
}

// TestSetNestedKey_TwoLevels verifies dot-notation creates nested maps.
func TestSetNestedKey_TwoLevels(t *testing.T) {
	m := map[string]any{}
	setNestedKey(m, "a.b", "val")
	sub, ok := m["a"].(map[string]any)
	if !ok {
		t.Fatalf("expected map at 'a', got %T", m["a"])
	}
	if sub["b"] != "val" {
		t.Errorf("expected b=val, got %v", sub["b"])
	}
}

// TestSetNestedKey_ThreeLevels verifies three-level dot-notation.
func TestSetNestedKey_ThreeLevels(t *testing.T) {
	m := map[string]any{}
	setNestedKey(m, "a.b.c", "deep")
	sub, _ := m["a"].(map[string]any)
	sub2, _ := sub["b"].(map[string]any)
	if sub2["c"] != "deep" {
		t.Errorf("expected c=deep, got %v", sub2["c"])
	}
}

// TestGetConfig_ReadsViper verifies that GetConfig reflects values set in Viper.
func TestGetConfig_ReadsViper(t *testing.T) {
	// Use a fresh Viper instance to avoid cross-test global state.
	viper.Reset()
	viper.Set("server.url", "http://getconfig-test:9999")
	viper.Set("server.token", "mytoken")
	viper.Set("company.default", "co-abc")
	t.Cleanup(func() { viper.Reset() })

	cfg := GetConfig()
	if cfg.ServerURL != "http://getconfig-test:9999" {
		t.Errorf("expected http://getconfig-test:9999, got %q", cfg.ServerURL)
	}
	if cfg.ServerToken != "mytoken" {
		t.Errorf("expected token=mytoken, got %q", cfg.ServerToken)
	}
	if cfg.CompanyDefault != "co-abc" {
		t.Errorf("expected company.default=co-abc, got %q", cfg.CompanyDefault)
	}
}

// TestConfigGetCmd verifies the config get subcommand reads from Viper.
func TestConfigGetCmd(t *testing.T) {
	viper.Reset()
	viper.Set("server.url", "http://test:1234")

	cmd := ConfigCmd()
	cmd.SetArgs([]string{"get", "server.url"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config get: %v", err)
	}
}

// TestConfigSetCmd verifies the config set subcommand persists values.
func TestConfigSetCmd(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, ".soksak", "config.json")
	os.MkdirAll(filepath.Dir(cfgFile), 0o700)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", origHome); viper.Reset() })

	viper.Reset()
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("json")

	cmd := ConfigCmd()
	cmd.SetArgs([]string{"set", "server.url", "http://set:5000"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set: %v", err)
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var raw map[string]any
	json.Unmarshal(data, &raw)
	server, _ := raw["server"].(map[string]any)
	if server["url"] != "http://set:5000" {
		t.Errorf("expected http://set:5000, got %v", server["url"])
	}
}
