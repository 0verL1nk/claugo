package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"holycode/internal/core"
)

type ReadTool struct{}

type readInput struct {
	Path string `json:"path"`
}

func (t ReadTool) Name() string { return "Read" }

func (t ReadTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        t.Name(),
		Description: "Read a file from the local workspace.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"],"additionalProperties":false}`),
	}
}

func (t ReadTool) Execute(_ context.Context, input []byte, state *State) (Result, error) {
	var req readInput
	if err := json.Unmarshal(input, &req); err != nil {
		return Result{}, err
	}
	content, err := os.ReadFile(req.Path)
	if err != nil {
		return Result{}, err
	}
	if state.ReadFiles == nil {
		state.ReadFiles = map[string]bool{}
	}
	state.ReadFiles[req.Path] = true
	return Result{Output: string(content)}, nil
}

func requireRead(state *State, path string) error {
	if state != nil && state.ReadFiles[path] {
		return nil
	}
	return fmt.Errorf("file must be read before it can be edited")
}
