package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds CLI runtime configuration loaded from disk + env overrides.
type Config struct {
	ServerURL      string
	ServerToken    string
	CompanyDefault string
}

// InitConfig loads ~/.soksak/config.json into Viper, then applies
// environment-variable overrides.  It is idempotent and safe to call
// multiple times.
func InitConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	cfgDir := filepath.Join(home, ".soksak")
	cfgFile := filepath.Join(cfgDir, "config.json")

	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("json")

	// Env overrides — e.g. SOKSAK_SERVER_URL
	viper.SetEnvPrefix("SOKSAK")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("server.url", "http://localhost:3100")
	viper.SetDefault("server.token", "")
	viper.SetDefault("company.default", "")

	// Read the file if it exists; ignore missing-file errors.
	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			// Ignore parse errors silently — let commands fail individually
			// when they try to build a client.
			_ = err
		}
	}
}

// GetConfig reads the current Viper state into a Config struct.
func GetConfig() Config {
	return Config{
		ServerURL:      viper.GetString("server.url"),
		ServerToken:    viper.GetString("server.token"),
		CompanyDefault: viper.GetString("company.default"),
	}
}

// configFilePath returns the path to ~/.soksak/config.json.
func configFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".soksak", "config.json")
}

// WriteConfigValue persists a key=value pair into the JSON config file.
func WriteConfigValue(key, value string) error {
	viper.Set(key, value)

	cfgFile := configFilePath()
	if err := os.MkdirAll(filepath.Dir(cfgFile), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Read existing JSON so we don't clobber unrelated keys.
	raw := map[string]any{}
	if data, err := os.ReadFile(cfgFile); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	setNestedKey(raw, key, value)

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(cfgFile, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// setNestedKey writes value into map using dot-notation key (e.g. "server.url").
func setNestedKey(m map[string]any, key, value string) {
	// For simplicity, top-level and two-segment paths are supported.
	for i, ch := range key {
		if ch == '.' {
			sub, ok := m[key[:i]].(map[string]any)
			if !ok {
				sub = map[string]any{}
				m[key[:i]] = sub
			}
			setNestedKey(sub, key[i+1:], value)
			return
		}
	}
	m[key] = value
}
