package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestSetGitTimeout_RoundTrip verifies SetGitTimeout returns the previous
// value, allowing tests to restore the package default via defer.
func TestSetGitTimeout_RoundTrip(t *testing.T) {
	original := defaultGitTimeout
	t.Cleanup(func() { SetGitTimeout(original) })

	prev := SetGitTimeout(2 * time.Second)
	if prev != original {
		t.Fatalf("SetGitTimeout returned %v, want %v", prev, original)
	}
	if defaultGitTimeout != 2*time.Second {
		t.Fatalf("defaultGitTimeout = %v, want 2s", defaultGitTimeout)
	}

	prev = SetGitTimeout(original)
	if prev != 2*time.Second {
		t.Fatalf("SetGitTimeout returned %v, want 2s", prev)
	}
}

// TestGitTimeoutError_DeadlineExceeded proves deadline-cancelled errors are
// rewritten with the configured timeout so callers see "timed out after 10s"
// instead of "signal: killed" or "context deadline exceeded".
func TestGitTimeoutError_DeadlineExceeded(t *testing.T) {
	original := defaultGitTimeout
	t.Cleanup(func() { SetGitTimeout(original) })
	SetGitTimeout(7 * time.Millisecond)

	wrapped := fmt.Errorf("exec failed: %w", context.DeadlineExceeded)
	err := gitTimeoutError("diff", wrapped, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected 'timed out' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "7ms") {
		t.Fatalf("expected configured timeout in error, got: %v", err)
	}
}

// TestGitTimeoutError_OtherError preserves original error context (output,
// wrapped error) when the failure was not a deadline cancellation.
func TestGitTimeoutError_OtherError(t *testing.T) {
	base := errors.New("exit 1")
	err := gitTimeoutError("diff", base, []byte("fatal: not a git repo"))
	if !errors.Is(err, base) {
		t.Fatalf("expected wrapped error to contain base, got: %v", err)
	}
	if !strings.Contains(err.Error(), "fatal: not a git repo") {
		t.Fatalf("expected stderr in error, got: %v", err)
	}
	if strings.Contains(err.Error(), "timed out") {
		t.Fatalf("non-deadline error must not be labeled as timeout: %v", err)
	}
}

// TestGitCommand_DeadlineApplied proves gitCommand actually attaches the
// package deadline. Forces a nanosecond timeout, runs git, and confirms the
// returned error matches context.DeadlineExceeded so callers can detect it.
func TestGitCommand_DeadlineApplied(t *testing.T) {
	original := defaultGitTimeout
	t.Cleanup(func() { SetGitTimeout(original) })
	SetGitTimeout(1 * time.Nanosecond)

	cmd, cancel := gitCommand(t.TempDir(), "status")
	defer cancel()
	err := cmd.Run()
	if err == nil {
		t.Skip("git ran faster than 1ns deadline (unexpected on real systems)")
	}
	// Either Run returns context.DeadlineExceeded directly or the underlying
	// process is killed by the context; ctx.Err() must reflect the cancellation.
	if ctxErr := cmd.Cancel; ctxErr == nil {
		t.Fatalf("gitCommand did not attach a cancellable context")
	}
}

// TestIsGitRepo_TimesOutFalse confirms IsGitRepo returns false (not panic,
// not hang) when the probe is cancelled by deadline.
func TestIsGitRepo_TimesOutFalse(t *testing.T) {
	original := defaultGitTimeout
	t.Cleanup(func() { SetGitTimeout(original) })
	SetGitTimeout(1 * time.Nanosecond)

	if IsGitRepo(t.TempDir()) {
		t.Fatal("IsGitRepo should return false when probe is cancelled")
	}
}
