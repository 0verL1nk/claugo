package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"holycode/internal/core"
)

type BashTool struct{}

type bashInput struct {
	Command string `json:"command"`
}

func (t BashTool) Name() string { return "Bash" }

func (t BashTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        t.Name(),
		Description: "Run a shell command in the current workspace.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"],"additionalProperties":false}`),
	}
}

func (t BashTool) Execute(ctx context.Context, input []byte, state *State) (Result, error) {
	var req bashInput
	if err := json.Unmarshal(input, &req); err != nil {
		return Result{}, err
	}
	if isUnsafeCommand(req.Command) {
		approved := state != nil && state.Approve != nil && state.Approve(req.Command)
		if !approved {
			return Result{Output: "command requires approval", IsError: true}, nil
		}
	}
	cmd := exec.CommandContext(ctx, "bash", "-lc", req.Command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Output: string(out), IsError: true}, nil
	}
	return Result{Output: string(out)}, nil
}

func isUnsafeCommand(command string) bool {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return false
	}
	unsafeFragments := []string{"rm ", "mv ", "cp ", "touch ", "mkdir ", ">", ">>"}
	for _, fragment := range unsafeFragments {
		if strings.Contains(trimmed, fragment) {
			return true
		}
	}
	return false
}
