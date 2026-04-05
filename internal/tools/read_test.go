package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestReadRegistersFileState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	writeTestFile(t, path, "hello")

	payload, err := json.Marshal(map[string]string{"path": path})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	state := &State{ReadFiles: map[string]bool{}}
	result, err := ReadTool{}.Execute(context.Background(), payload, state)
	if err != nil {
		t.Fatalf("read execute: %v", err)
	}
	if result.Output != "hello" {
		t.Fatalf("expected file content %q, got %q", "hello", result.Output)
	}
	if !state.ReadFiles[path] {
		t.Fatalf("expected %q to be marked as read", path)
	}
}
