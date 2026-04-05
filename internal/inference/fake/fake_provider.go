package fake

import (
	"context"

	"holycode/internal/inference"
)

type Provider struct {
	ProviderName string
	DescriptorFn func(model string, baseURL string) inference.ModelDescriptor
	RunTurnFn    func(ctx context.Context, descriptor inference.ModelDescriptor, apiKey string, req inference.TurnRequest) (<-chan inference.Event, error)
}

func (p *Provider) Name() string {
	if p.ProviderName == "" {
		return "fake"
	}
	return p.ProviderName
}

func (p *Provider) Descriptor(model string, baseURL string) inference.ModelDescriptor {
	if p.DescriptorFn != nil {
		return p.DescriptorFn(model, baseURL)
	}
	return inference.ModelDescriptor{
		ProviderName: p.Name(),
		Model:        model,
		BaseURL:      baseURL,
		Capabilities: inference.Capabilities{SupportsToolCalls: true, SupportsConversation: true},
	}
}

func (p *Provider) RunTurn(ctx context.Context, descriptor inference.ModelDescriptor, apiKey string, req inference.TurnRequest) (<-chan inference.Event, error) {
	if p.RunTurnFn != nil {
		return p.RunTurnFn(ctx, descriptor, apiKey, req)
	}
	ch := make(chan inference.Event, 1)
	ch <- inference.Event{
		Type:         inference.EventTypeCompleted,
		StopReason:   inference.StopReasonCompleted,
		ProviderName: p.Name(),
	}
	close(ch)
	return ch, nil
}
