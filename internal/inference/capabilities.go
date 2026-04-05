package inference

type Capabilities struct {
	SupportsToolCalls        bool
	SupportsToolArgStreaming bool
	SupportsStructuredOutput bool
	SupportsConversation     bool
}

type ModelDescriptor struct {
	ProviderName string
	Model        string
	BaseURL      string
	Capabilities Capabilities
}
