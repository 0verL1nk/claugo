package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBashRejectsUnsafeCommandWithoutApproval(t *testing.T) {
	payload, err := json.Marshal(map[string]string{"command": "touch /tmp/holy-test"})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	result, err := BashTool{}.Execute(context.Background(), payload, &State{
		Approve: func(string) bool { return false },
	})
	if err != nil {
		t.Fatalf("bash execute returned unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected bash result to be marked as error")
	}
	if result.Output == "" {
		t.Fatal("expected bash rejection message")
	}
}
