package core

import "encoding/json"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type ContentBlockType string

const (
	ContentBlockText ContentBlockType = "text"
)

type ContentBlock struct {
	Type ContentBlockType
	Text string
}

type Message struct {
	Role    Role
	Content []ContentBlock
}

type ToolDefinition struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}
