package config

import (
	"strings"
	"testing"
)

func TestUpstreamAPIKeyEnvConfig(t *testing.T) {
	t.Setenv("LLMSNAP_UPSTREAM_CONFIG_TEST_KEY", "backend-test-secret")

	valid := `models:
  managed:
    cmd: echo dummy
    proxy: http://127.0.0.1:8101
    upstreamApiKeyEnv: LLMSNAP_UPSTREAM_CONFIG_TEST_KEY
`
	cfg, err := LoadConfigFromReader(strings.NewReader(valid))
	if err != nil {
		t.Fatalf("valid managed model config: %v", err)
	}
	if got := cfg.Models["managed"].UpstreamAPIKeyEnv; got != "LLMSNAP_UPSTREAM_CONFIG_TEST_KEY" {
		t.Fatalf("upstream key env name = %q", got)
	}

	t.Setenv("LLMSNAP_UPSTREAM_CONFIG_TEST_KEY", "")
	if _, err := LoadConfigFromReader(strings.NewReader(valid)); err == nil ||
		!strings.Contains(err.Error(), "LLMSNAP_UPSTREAM_CONFIG_TEST_KEY") {
		t.Fatalf("missing or empty backend key must fail without exposing values, got %v", err)
	}

	invalid := strings.Replace(valid, "LLMSNAP_UPSTREAM_CONFIG_TEST_KEY", "not-a-valid-env-var", 1)
	if _, err := LoadConfigFromReader(strings.NewReader(invalid)); err == nil ||
		!strings.Contains(err.Error(), "invalid upstreamApiKeyEnv") {
		t.Fatalf("invalid variable name must fail, got %v", err)
	}

	unsecured := `models:
  unsecured:
    cmd: echo dummy
    proxy: http://127.0.0.1:8101
`
	if _, err := LoadConfigFromReader(strings.NewReader(unsecured)); err != nil {
		t.Fatalf("backends with no API key remain supported: %v", err)
	}
}
