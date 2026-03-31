package proxy

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type Result struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
	Duration string `json:"duration"`
}

type Executor struct {
	timeout         time.Duration
	allowedCommands map[string]bool
}

func NewExecutor(timeoutSec int, allowed []string) *Executor {
	m := make(map[string]bool, len(allowed))
	for _, cmd := range allowed {
		m[cmd] = true
	}
	return &Executor{
		timeout:         time.Duration(timeoutSec) * time.Second,
		allowedCommands: m,
	}
}

// Run executes a command and returns the result.
// Only allows commands from a predefined allowlist for security.
func (e *Executor) Run(ctx context.Context, name string, args ...string) (*Result, error) {
	if !e.isAllowed(name) {
		return nil, fmt.Errorf("command %q is not in the allowlist", name)
	}

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Duration: time.Since(start).String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("executor.Run: %w", err)
		}
	}

	return result, nil
}

// RunStream executes a command and streams stdout line by line via a channel.
func (e *Executor) RunStream(ctx context.Context, name string, args ...string) (<-chan string, <-chan error) {
	lines := make(chan string, 100)
	errs := make(chan error, 1)

	if !e.isAllowed(name) {
		go func() {
			errs <- fmt.Errorf("command %q is not in the allowlist", name)
			close(lines)
			close(errs)
		}()
		return lines, errs
	}

	go func() {
		defer close(lines)
		defer close(errs)

		ctx, cancel := context.WithTimeout(ctx, e.timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, name, args...)
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			errs <- err
			return
		}

		if err := cmd.Start(); err != nil {
			errs <- err
			return
		}

		buf := make([]byte, 4096)
		for {
			n, err := pipe.Read(buf)
			if n > 0 {
				lines <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}

		if err := cmd.Wait(); err != nil {
			errs <- err
		}
	}()

	return lines, errs
}

func (e *Executor) isAllowed(name string) bool {
	return e.allowedCommands[name]
}
