package inference

import "fmt"

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	reg := &Registry{providers: map[string]Provider{}}
	for _, provider := range providers {
		reg.providers[provider.Name()] = provider
	}
	return reg
}

func (r *Registry) Lookup(name string) (Provider, error) {
	if r == nil {
		return nil, fmt.Errorf("provider registry is nil")
	}
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q is not registered", name)
	}
	return provider, nil
}
