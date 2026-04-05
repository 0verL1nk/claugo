package tools

import "context"

import "holycode/internal/core"

type Tool interface {
	Name() string
	Definition() core.ToolDefinition
	Execute(ctx context.Context, input []byte, state *State) (Result, error)
}

type Result struct {
	Output  string
	IsError bool
}

type State struct {
	ReadFiles map[string]bool
	Approve   func(command string) bool
}
