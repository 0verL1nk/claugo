package core

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigFromArgsAnthropicReadsAPIKeyFromEnv(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	cfg, err := ConfigFromArgs([]string{"--provider", "anthropic", "--model", "claude-test", "hello"}, bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("ConfigFromArgs returned error: %v", err)
	}
	if cfg.APIKey != "test-key" {
		t.Fatalf("expected API key from env, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.anthropic.com" {
		t.Fatalf("expected default Anthropic base URL, got %q", cfg.BaseURL)
	}
}

func TestConfigFromArgsAnthropicRequiresModel(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	_, err := ConfigFromArgs([]string{"--provider", "anthropic", "hello"}, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatal("expected error for missing Anthropic model")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Fatalf("expected model error, got %v", err)
	}
}

func TestConfigFromArgsReadsProjectSettingsFile(t *testing.T) {
	home := t.TempDir()
	workdir := t.TempDir()
	t.Setenv("HOME", home)

	settingsDir := filepath.Join(workdir, ".holy")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	settingsPath := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"provider":"anthropic","model":"claude-from-project","api_key":"project-key","base_url":"https://project.example"}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldwd)
	})

	cfg, err := ConfigFromArgs([]string{"hello"}, bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("ConfigFromArgs returned error: %v", err)
	}
	if cfg.ProviderName != "anthropic" {
		t.Fatalf("expected provider from settings, got %q", cfg.ProviderName)
	}
	if cfg.Model != "claude-from-project" {
		t.Fatalf("expected model from settings, got %q", cfg.Model)
	}
	if cfg.APIKey != "project-key" {
		t.Fatalf("expected api key from settings, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://project.example" {
		t.Fatalf("expected base URL from settings, got %q", cfg.BaseURL)
	}
}

func TestConfigFromArgsFlagsOverrideSettingsFile(t *testing.T) {
	home := t.TempDir()
	workdir := t.TempDir()
	t.Setenv("HOME", home)

	settingsDir := filepath.Join(workdir, ".holy")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	settingsPath := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"provider":"anthropic","model":"claude-from-project","api_key":"project-key","base_url":"https://project.example"}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldwd)
	})

	cfg, err := ConfigFromArgs([]string{"--provider", "anthropic", "--model", "claude-from-flag", "--api-key", "flag-key", "--base-url", "https://flag.example", "hello"}, bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("ConfigFromArgs returned error: %v", err)
	}
	if cfg.Model != "claude-from-flag" {
		t.Fatalf("expected model from flag, got %q", cfg.Model)
	}
	if cfg.APIKey != "flag-key" {
		t.Fatalf("expected api key from flag, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://flag.example" {
		t.Fatalf("expected base URL from flag, got %q", cfg.BaseURL)
	}
}

func TestConfigFromArgsOpenAIResponsesReadsAPIKeyFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai-test-key")

	cfg, err := ConfigFromArgs([]string{"--provider", "openai-responses", "--model", "gpt-5", "hello"}, bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("ConfigFromArgs returned error: %v", err)
	}
	if cfg.APIKey != "openai-test-key" {
		t.Fatalf("expected API key from env, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected default OpenAI base URL, got %q", cfg.BaseURL)
	}
}

func TestConfigFromArgsOpenAICompatibleRequiresBaseURL(t *testing.T) {
	t.Setenv("OPENAI_COMPATIBLE_API_KEY", "compatible-key")

	_, err := ConfigFromArgs([]string{"--provider", "openai-compatible", "--model", "llama-test", "hello"}, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatal("expected error for missing OpenAI-compatible base URL")
	}
	if !strings.Contains(err.Error(), "base") {
		t.Fatalf("expected base URL error, got %v", err)
	}
}

func TestConfigFromArgsOpenAICompatibleReadsAPIKeyFromEnv(t *testing.T) {
	t.Setenv("OPENAI_COMPATIBLE_API_KEY", "compatible-key")

	cfg, err := ConfigFromArgs([]string{"--provider", "openai-compatible", "--model", "llama-test", "--base-url", "https://compat.example/v1", "hello"}, bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("ConfigFromArgs returned error: %v", err)
	}
	if cfg.APIKey != "compatible-key" {
		t.Fatalf("expected API key from env, got %q", cfg.APIKey)
	}
}
