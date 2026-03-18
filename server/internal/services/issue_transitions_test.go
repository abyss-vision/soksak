package services

import (
	"testing"
)

// TestIssueTransitions covers every path in ValidateTransition, achieving 100% coverage.
func TestIssueTransitions(t *testing.T) {
	t.Run("unknown from status", func(t *testing.T) {
		err := ValidateTransition("unknown", "todo")
		if err == nil {
			t.Fatal("expected error for unknown from status, got nil")
		}
	})

	t.Run("backlog to todo", func(t *testing.T) {
		if err := ValidateTransition("backlog", "todo"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("backlog to in_progress", func(t *testing.T) {
		if err := ValidateTransition("backlog", "in_progress"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("backlog to cancelled", func(t *testing.T) {
		if err := ValidateTransition("backlog", "cancelled"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("backlog to done is invalid", func(t *testing.T) {
		err := ValidateTransition("backlog", "done")
		if err == nil {
			t.Fatal("expected error transitioning backlog -> done")
		}
	})

	t.Run("todo to backlog", func(t *testing.T) {
		if err := ValidateTransition("todo", "backlog"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("todo to in_progress", func(t *testing.T) {
		if err := ValidateTransition("todo", "in_progress"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("todo to cancelled", func(t *testing.T) {
		if err := ValidateTransition("todo", "cancelled"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("todo to done is invalid", func(t *testing.T) {
		err := ValidateTransition("todo", "done")
		if err == nil {
			t.Fatal("expected error transitioning todo -> done")
		}
	})

	t.Run("in_progress to in_review", func(t *testing.T) {
		if err := ValidateTransition("in_progress", "in_review"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_progress to blocked", func(t *testing.T) {
		if err := ValidateTransition("in_progress", "blocked"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_progress to done", func(t *testing.T) {
		if err := ValidateTransition("in_progress", "done"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_progress to cancelled", func(t *testing.T) {
		if err := ValidateTransition("in_progress", "cancelled"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_progress to backlog is invalid", func(t *testing.T) {
		err := ValidateTransition("in_progress", "backlog")
		if err == nil {
			t.Fatal("expected error transitioning in_progress -> backlog")
		}
	})

	t.Run("in_review to in_progress", func(t *testing.T) {
		if err := ValidateTransition("in_review", "in_progress"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_review to done", func(t *testing.T) {
		if err := ValidateTransition("in_review", "done"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_review to cancelled", func(t *testing.T) {
		if err := ValidateTransition("in_review", "cancelled"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("in_review to backlog is invalid", func(t *testing.T) {
		err := ValidateTransition("in_review", "backlog")
		if err == nil {
			t.Fatal("expected error transitioning in_review -> backlog")
		}
	})

	t.Run("blocked to in_progress", func(t *testing.T) {
		if err := ValidateTransition("blocked", "in_progress"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("blocked to cancelled", func(t *testing.T) {
		if err := ValidateTransition("blocked", "cancelled"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("blocked to done is invalid", func(t *testing.T) {
		err := ValidateTransition("blocked", "done")
		if err == nil {
			t.Fatal("expected error transitioning blocked -> done")
		}
	})

	t.Run("done is terminal", func(t *testing.T) {
		err := ValidateTransition("done", "backlog")
		if err == nil {
			t.Fatal("expected error transitioning from terminal done")
		}
	})

	t.Run("done to done is terminal", func(t *testing.T) {
		err := ValidateTransition("done", "done")
		if err == nil {
			t.Fatal("expected error transitioning done -> done (terminal)")
		}
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		err := ValidateTransition("cancelled", "backlog")
		if err == nil {
			t.Fatal("expected error transitioning from terminal cancelled")
		}
	})

	t.Run("cancelled to cancelled is terminal", func(t *testing.T) {
		err := ValidateTransition("cancelled", "cancelled")
		if err == nil {
			t.Fatal("expected error transitioning cancelled -> cancelled (terminal)")
		}
	})
}
