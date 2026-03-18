package services

import (
	"fmt"

	"abyss-view/internal/domain"
)

// validTransitions maps each issue status to the set of statuses it can transition to.
var validTransitions = map[domain.IssueStatus][]domain.IssueStatus{
	domain.IssueStatusBacklog:    {domain.IssueStatusTodo, domain.IssueStatusInProgress, domain.IssueStatusCancelled},
	domain.IssueStatusTodo:       {domain.IssueStatusBacklog, domain.IssueStatusInProgress, domain.IssueStatusCancelled},
	domain.IssueStatusInProgress: {domain.IssueStatusInReview, domain.IssueStatusBlocked, domain.IssueStatusDone, domain.IssueStatusCancelled},
	domain.IssueStatusInReview:   {domain.IssueStatusInProgress, domain.IssueStatusDone, domain.IssueStatusCancelled},
	domain.IssueStatusBlocked:    {domain.IssueStatusInProgress, domain.IssueStatusCancelled},
	domain.IssueStatusDone:       {},
	domain.IssueStatusCancelled:  {},
}

// ValidateTransition returns an error if transitioning from → to is not allowed.
func ValidateTransition(from, to string) error {
	fromStatus := domain.IssueStatus(from)
	toStatus := domain.IssueStatus(to)

	allowed, known := validTransitions[fromStatus]
	if !known {
		return fmt.Errorf("unknown issue status: %q", from)
	}

	for _, s := range allowed {
		if s == toStatus {
			return nil
		}
	}

	if len(allowed) == 0 {
		return fmt.Errorf("status %q is terminal and cannot be changed", from)
	}

	return fmt.Errorf("cannot transition issue from %q to %q", from, to)
}
