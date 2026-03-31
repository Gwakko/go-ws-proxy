package proxy

import (
	"context"
	"strings"
	"testing"
)

func TestAllowedCommandRunsSuccessfully(t *testing.T) {
	e := NewExecutor(5)
	result, err := e.Run(context.Background(), "hostname")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Stdout == "" {
		t.Fatal("expected non-empty stdout from hostname")
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestDisallowedCommandReturnsError(t *testing.T) {
	e := NewExecutor(5)
	_, err := e.Run(context.Background(), "rm")
	if err == nil {
		t.Fatal("expected error for disallowed command, got nil")
	}
	if !strings.Contains(err.Error(), "not in the allowlist") {
		t.Fatalf("expected allowlist error, got: %v", err)
	}
}

func TestRunReturnsStdoutContent(t *testing.T) {
	e := NewExecutor(5)
	result, err := e.Run(context.Background(), "whoami")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	trimmed := strings.TrimSpace(result.Stdout)
	if trimmed == "" {
		t.Fatal("expected whoami to produce output")
	}
}

func TestRunCapturesExitCodeOnFailure(t *testing.T) {
	e := NewExecutor(5)
	result, err := e.Run(context.Background(), "ls", "/nonexistent_path_that_should_not_exist")
	if err != nil {
		t.Fatalf("expected no wrapper error for non-zero exit, got %v", err)
	}
	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code for ls on missing path")
	}
}

func TestRunStreamSendsOutputThroughChannel(t *testing.T) {
	e := NewExecutor(5)
	lines, errs := e.RunStream(context.Background(), "hostname")

	var output strings.Builder
	for line := range lines {
		output.WriteString(line)
	}

	if err := <-errs; err != nil {
		t.Fatalf("expected no error from stream, got %v", err)
	}

	if strings.TrimSpace(output.String()) == "" {
		t.Fatal("expected streamed output from hostname")
	}
}

func TestRunStreamDisallowedCommand(t *testing.T) {
	e := NewExecutor(5)
	lines, errs := e.RunStream(context.Background(), "rm")

	for range lines {
		// drain
	}

	err := <-errs
	if err == nil {
		t.Fatal("expected error for disallowed command in stream")
	}
	if !strings.Contains(err.Error(), "not in the allowlist") {
		t.Fatalf("expected allowlist error, got: %v", err)
	}
}
