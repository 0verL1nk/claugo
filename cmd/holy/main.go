package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"holycode/internal/api"
	"holycode/internal/core"
	"holycode/internal/inference"
	fakeprovider "holycode/internal/inference/fake"
	"holycode/internal/providers/anthropic"
	"holycode/internal/providers/openaicompat"
	"holycode/internal/providers/openairesponses"
	"holycode/internal/query"
	"holycode/internal/tools"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg, err := core.ConfigFromArgs(args, stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	runtime := api.NewRuntime(newProviderRegistry())
	result, err := query.Run(context.Background(), runtime, tools.NewRegistry(
		tools.ReadTool{},
		tools.EditTool{},
		tools.BashTool{},
	), cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintln(stdout, result.Text)
	return 0
}

func newProviderRegistry() *inference.Registry {
	return inference.NewRegistry(
		&fakeprovider.Provider{
			RunTurnFn: func(_ context.Context, _ inference.ModelDescriptor, _ string, req inference.TurnRequest) (<-chan inference.Event, error) {
				ch := make(chan inference.Event, 2)
				ch <- inference.Event{Type: inference.EventTypeTextDelta, TextDelta: req.Prompt}
				ch <- inference.Event{Type: inference.EventTypeCompleted, StopReason: inference.StopReasonCompleted}
				close(ch)
				return ch, nil
			},
		},
		&anthropic.Provider{Client: http.DefaultClient},
		&openairesponses.Provider{Client: http.DefaultClient},
		&openaicompat.Provider{Client: http.DefaultClient},
	)
}
