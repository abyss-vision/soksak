package domain

import "testing"

func TestValidateIssuePrefix(t *testing.T) {
	valid := []string{"AB", "ABC", "ABCDE", "ABCDEFGHIJ"} // 2-10 uppercase
	for _, p := range valid {
		if err := ValidateIssuePrefix(p); err != nil {
			t.Errorf("ValidateIssuePrefix(%q): unexpected error: %v", p, err)
		}
	}

	invalid := []string{"", "A", "a", "Ab", "ABCDEFGHIJK", "AB1", "1AB"}
	for _, p := range invalid {
		if err := ValidateIssuePrefix(p); err == nil {
			t.Errorf("ValidateIssuePrefix(%q): expected error, got nil", p)
		}
	}
}

func TestValidateUUID(t *testing.T) {
	valid := []string{
		"123e4567-e89b-12d3-a456-426614174000",
		"00000000-0000-0000-0000-000000000000",
		"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", // uppercase accepted via ToLower
	}
	for _, id := range valid {
		if err := ValidateUUID(id); err != nil {
			t.Errorf("ValidateUUID(%q): unexpected error: %v", id, err)
		}
	}

	invalid := []string{
		"",
		"not-a-uuid",
		"123e4567-e89b-12d3-a456-42661417400", // too short
		"123e4567-e89b-12d3-a456-4266141740000", // too long
		"123e4567-e89b-12d3-a456_426614174000", // wrong separator
	}
	for _, id := range invalid {
		if err := ValidateUUID(id); err == nil {
			t.Errorf("ValidateUUID(%q): expected error, got nil", id)
		}
	}
}

func TestValidateCompanyStatus(t *testing.T) {
	valid := []string{"active", "paused", "deleted"}
	for _, s := range valid {
		if err := ValidateCompanyStatus(s); err != nil {
			t.Errorf("ValidateCompanyStatus(%q): unexpected error: %v", s, err)
		}
	}
	if err := ValidateCompanyStatus("unknown"); err == nil {
		t.Error("ValidateCompanyStatus(\"unknown\"): expected error, got nil")
	}
}

func TestValidateIssueStatus(t *testing.T) {
	valid := []string{"backlog", "todo", "in_progress", "done", "cancelled"}
	for _, s := range valid {
		if err := ValidateIssueStatus(s); err != nil {
			t.Errorf("ValidateIssueStatus(%q): unexpected error: %v", s, err)
		}
	}
	if err := ValidateIssueStatus("invalid"); err == nil {
		t.Error("ValidateIssueStatus(\"invalid\"): expected error, got nil")
	}
}

func TestValidatePriority(t *testing.T) {
	valid := []string{"urgent", "high", "medium", "low", "none"}
	for _, s := range valid {
		if err := ValidatePriority(s); err != nil {
			t.Errorf("ValidatePriority(%q): unexpected error: %v", s, err)
		}
	}
	if err := ValidatePriority("critical"); err == nil {
		t.Error("ValidatePriority(\"critical\"): expected error, got nil")
	}
}

func TestValidateAgentStatus(t *testing.T) {
	valid := []string{"idle", "running", "paused", "deleted"}
	for _, s := range valid {
		if err := ValidateAgentStatus(s); err != nil {
			t.Errorf("ValidateAgentStatus(%q): unexpected error: %v", s, err)
		}
	}
	if err := ValidateAgentStatus("stopped"); err == nil {
		t.Error("ValidateAgentStatus(\"stopped\"): expected error, got nil")
	}
}
