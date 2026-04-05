package core

import "fmt"

type ErrorKind string

const (
	ErrorKindConfig    ErrorKind = "config"
	ErrorKindAuth      ErrorKind = "auth"
	ErrorKindRateLimit ErrorKind = "rate_limit"
	ErrorKindProvider  ErrorKind = "provider"
	ErrorKindTool      ErrorKind = "tool"
	ErrorKindOther     ErrorKind = "other"
)

type RuntimeError struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *RuntimeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
}

func (e *RuntimeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
