package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var issuePrefixRegex = regexp.MustCompile(`^[A-Z]{2,10}$`)
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidateIssuePrefix checks that an issue prefix is valid (2-10 uppercase letters).
func ValidateIssuePrefix(prefix string) error {
	if !issuePrefixRegex.MatchString(prefix) {
		return fmt.Errorf("issue prefix must be 2-10 uppercase letters, got %q", prefix)
	}
	return nil
}

// ValidateUUID checks that a string is a valid UUID format.
func ValidateUUID(id string) error {
	if !uuidRegex.MatchString(strings.ToLower(id)) {
		return fmt.Errorf("invalid UUID format: %q", id)
	}
	return nil
}

// ValidateCompanyStatus checks that a company status value is known.
func ValidateCompanyStatus(s string) error {
	switch CompanyStatus(s) {
	case CompanyStatusActive, CompanyStatusPaused, CompanyStatusDeleted:
		return nil
	}
	return fmt.Errorf("unknown company status %q", s)
}

// ValidateIssueStatus checks that an issue status value is known.
func ValidateIssueStatus(s string) error {
	switch IssueStatus(s) {
	case IssueStatusBacklog, IssueStatusTodo, IssueStatusInProgress, IssueStatusDone, IssueStatusCancelled:
		return nil
	}
	return fmt.Errorf("unknown issue status %q", s)
}

// ValidatePriority checks that a priority value is known.
func ValidatePriority(s string) error {
	switch Priority(s) {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
		return nil
	}
	return fmt.Errorf("unknown priority %q", s)
}

// ValidateAgentStatus checks that an agent status value is known.
func ValidateAgentStatus(s string) error {
	switch AgentStatus(s) {
	case AgentStatusIdle, AgentStatusRunning, AgentStatusPaused, AgentStatusDeleted:
		return nil
	}
	return fmt.Errorf("unknown agent status %q", s)
}
