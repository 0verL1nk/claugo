package tools

import (
	"fmt"
	"sort"

	"holycode/internal/core"
)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry(tools ...Tool) *Registry {
	reg := &Registry{tools: map[string]Tool{}}
	for _, tool := range tools {
		reg.tools[tool.Name()] = tool
	}
	return reg
}

func (r *Registry) Lookup(name string) (Tool, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool %q is not registered", name)
	}
	return tool, nil
}

func (r *Registry) Definitions() []core.ToolDefinition {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	defs := make([]core.ToolDefinition, 0, len(names))
	for _, name := range names {
		defs = append(defs, r.tools[name].Definition())
	}
	return defs
}
