package exec

import (
	"context"
	"testing"
)

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		r    Result
		want bool
	}{
		{"transient stderr", Result{Stderr: "the resource doesn't exist yet"}, true},
		{"throttling", Result{Stdout: "TooManyRequests"}, true},
		{"conflict", Result{Stderr: "Conflict: AnotherOperationInProgress"}, true},
		{"hard failure", Result{Stderr: "AuthorizationFailed"}, false},
		{"clean", Result{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isRetryable(c.r); got != c.want {
				t.Fatalf("isRetryable(%+v) = %v, want %v", c.r, got, c.want)
			}
		})
	}
}

func TestRunCapturesExit(t *testing.T) {
	ctx := context.Background()

	ok := Run(ctx, nil, "true")
	if ok.ExitCode != 0 {
		t.Fatalf("`true` exit = %d", ok.ExitCode)
	}

	bad := Run(ctx, nil, "false")
	if bad.ExitCode == 0 {
		t.Fatalf("`false` should exit non-zero")
	}
}

func TestRunWithRetriesStopsOnNonRetryable(t *testing.T) {
	ctx := context.Background()
	// `false` fails with no retryable marker, so it must not be retried.
	_, err := RunWithRetries(ctx, Options{MaxAttempts: 3, DelaySeconds: 1}, "false")
	if err == nil {
		t.Fatalf("expected error from failing command")
	}
}
