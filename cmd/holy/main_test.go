package main

import (
	"bytes"
	"testing"
)

func TestRunPrintsHeadlessResponseFromArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"--provider", "fake", "hello world"}, bytes.NewBuffer(nil), stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if stdout.String() != "hello world\n" {
		t.Fatalf("expected stdout %q, got %q", "hello world\n", stdout.String())
	}
}

func TestRunAcceptsPromptFromStdin(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"--provider", "fake"}, bytes.NewBufferString("from stdin"), stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if stdout.String() != "from stdin\n" {
		t.Fatalf("expected stdout %q, got %q", "from stdin\n", stdout.String())
	}
}

func TestRunReturnsNonZeroForUnknownProvider(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"--provider", "missing", "hello"}, bytes.NewBuffer(nil), stdout, stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if stderr.Len() == 0 {
		t.Fatal("expected error text on stderr")
	}
}

func TestRunReturnsNonZeroForAnthropicWithoutModel(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"--provider", "anthropic", "hello"}, bytes.NewBuffer(nil), stdout, stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if stderr.Len() == 0 {
		t.Fatal("expected error text on stderr")
	}
}

func TestRunReturnsNonZeroForAnthropicWithoutAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"--provider", "anthropic", "--model", "claude-test", "hello"}, bytes.NewBuffer(nil), stdout, stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if stderr.Len() == 0 {
		t.Fatal("expected error text on stderr")
	}
}

func TestNewProviderRegistryRegistersOpenAIProviderFamilies(t *testing.T) {
	registry := newProviderRegistry()
	if _, err := registry.Lookup("openai-responses"); err != nil {
		t.Fatalf("expected openai-responses provider, got error: %v", err)
	}
	if _, err := registry.Lookup("openai-compatible"); err != nil {
		t.Fatalf("expected openai-compatible provider, got error: %v", err)
	}
}
