package api

import (
	"context"

	"holycode/internal/core"
	"holycode/internal/inference"
)

type Runtime struct {
	registry *inference.Registry
}

func NewRuntime(registry *inference.Registry) *Runtime {
	return &Runtime{registry: registry}
}

func (r *Runtime) ResolveDescriptor(cfg core.Config) (inference.ModelDescriptor, error) {
	provider, err := r.registry.Lookup(cfg.ProviderName)
	if err != nil {
		return inference.ModelDescriptor{}, err
	}
	return provider.Descriptor(cfg.Model, cfg.BaseURL), nil
}

func (r *Runtime) RunTurn(ctx context.Context, cfg core.Config, req inference.TurnRequest) (<-chan inference.Event, inference.ModelDescriptor, error) {
	provider, err := r.registry.Lookup(cfg.ProviderName)
	if err != nil {
		return nil, inference.ModelDescriptor{}, err
	}
	descriptor := provider.Descriptor(cfg.Model, cfg.BaseURL)
	if len(req.ToolDefinitions) > 0 && !descriptor.Capabilities.SupportsToolCalls {
		return nil, descriptor, &core.RuntimeError{
			Kind:    core.ErrorKindProvider,
			Message: "selected model does not support tool calls",
		}
	}
	if len(req.AssistantTurns) > 0 && !descriptor.Capabilities.SupportsConversation {
		return nil, descriptor, &core.RuntimeError{
			Kind:    core.ErrorKindProvider,
			Message: "selected model does not support tool continuation",
		}
	}
	events, err := provider.RunTurn(ctx, descriptor, cfg.APIKey, req)
	if err != nil {
		return nil, descriptor, err
	}
	return events, descriptor, nil
}
