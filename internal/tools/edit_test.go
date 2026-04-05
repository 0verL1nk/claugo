package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEditRejectsExistingFileWithoutPriorRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	writeTestFile(t, path, "before")

	payload, err := json.Marshal(map[string]string{
		"path":    path,
		"content": "after",
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	result, err := EditTool{}.Execute(context.Background(), payload, &State{ReadFiles: map[string]bool{}})
	if err != nil {
		t.Fatalf("edit execute returned unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected edit result to be marked as error")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file after edit: %v", err)
	}
	if string(content) != "before" {
		t.Fatalf("expected original content to remain, got %q", string(content))
	}
}
