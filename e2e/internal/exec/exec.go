// Package exec is the runner-agnostic exec + bounded-retry helper, mirroring
// datadog-ci's e2e/helpers/exec.ts (execPromiseWithRetries). It retries the
// cloud, not the assertions: only transient provider errors are retried, on a
// bounded budget. A real failure surfaces immediately.
package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result captures a finished command's output and exit code.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// retryablePatterns are transient cloud-provider errors safe to retry,
// mirroring the reference impl's RETRYABLE_PATTERNS.
var retryablePatterns = []string{
	"GatewayTimeout",
	"RestError",
	"Operation was canceled",
	"ETIMEDOUT",
	"ECONNRESET",
	"doesn't exist",
	"Conflict",
	"TooManyRequests",
	"ABORTED",
	"DEADLINE_EXCEEDED",
	"INTERNAL",
	"RESOURCE_EXHAUSTED",
	"UNAVAILABLE",
	"temporarily unavailable",
	"AnotherOperationInProgress",
	"ServiceUnavailable",
}

func isRetryable(r Result) bool {
	out := r.Stdout + " " + r.Stderr
	for _, p := range retryablePatterns {
		if strings.Contains(out, p) {
			return true
		}
	}
	return false
}

// Options tunes the retry budget. Zero values fall back to the reference
// defaults (3 attempts, 5s delay).
type Options struct {
	MaxAttempts  int
	DelaySeconds int
	Env          []string // extra "KEY=VALUE" entries appended to os.Environ()
}

// Run executes a command once and returns its Result (never errors on non-zero
// exit; inspect Result.ExitCode).
func Run(ctx context.Context, env []string, name string, args ...string) Result {
	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = env
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := Result{
		ExitCode: 0,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		} else {
			res.ExitCode = 1
			if res.Stderr == "" {
				res.Stderr = err.Error()
			}
		}
	}
	return res
}

// RunWithRetries runs a command, retrying only on transient errors up to the
// configured budget. It returns the final Result and an error if the command
// never succeeded.
func RunWithRetries(ctx context.Context, opts Options, name string, args ...string) (Result, error) {
	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	delay := opts.DelaySeconds
	if delay <= 0 {
		delay = 5
	}

	var last Result
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		last = Run(ctx, opts.Env, name, args...)
		if last.ExitCode == 0 {
			return last, nil
		}
		if attempt < maxAttempts && isRetryable(last) {
			select {
			case <-ctx.Done():
				return last, ctx.Err()
			case <-time.After(time.Duration(delay) * time.Second):
			}
			continue
		}
		break
	}
	return last, fmt.Errorf("command %q failed after retries (exit %d): %s", name, last.ExitCode, last.Stderr)
}
