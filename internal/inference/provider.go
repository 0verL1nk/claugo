package inference

import "context"

type Provider interface {
	Name() string
	Descriptor(model string, baseURL string) ModelDescriptor
	RunTurn(ctx context.Context, descriptor ModelDescriptor, apiKey string, req TurnRequest) (<-chan Event, error)
}
