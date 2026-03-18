package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"soksak/internal/middleware"
)

// DashboardService handles dashboard and sidebar badge aggregations.
type DashboardService struct {
	db *sqlx.DB
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(db *sqlx.DB) *DashboardService {
	return &DashboardService{db: db}
}

// AgentStatusCounts holds counts of agents grouped by status.
type AgentStatusCounts struct {
	Active  int `json:"active"`
	Running int `json:"running"`
	Paused  int `json:"paused"`
	Error   int `json:"error"`
}

// IssueCounts holds counts of issues grouped by status.
type IssueCounts struct {
	Open       int `json:"open"`
	InProgress int `json:"inProgress"`
	Blocked    int `json:"blocked"`
	Done       int `json:"done"`
}

// CostInfo holds monthly cost summary.
type CostInfo struct {
	MonthSpendCents        int     `json:"monthSpendCents"`
	MonthBudgetCents       int     `json:"monthBudgetCents"`
	MonthUtilizationPercent float64 `json:"monthUtilizationPercent"`
}

// BudgetInfo holds budget incident summary.
type BudgetInfo struct {
	ActiveIncidents  int `json:"activeIncidents"`
	PendingApprovals int `json:"pendingApprovals"`
	PausedAgents     int `json:"pausedAgents"`
	PausedProjects   int `json:"pausedProjects"`
}

// DashboardSummary is the full dashboard response.
type DashboardSummary struct {
	CompanyUUID      string            `json:"companyUuid"`
	Agents           AgentStatusCounts `json:"agents"`
	Tasks            IssueCounts       `json:"tasks"`
	Costs            CostInfo          `json:"costs"`
	PendingApprovals int               `json:"pendingApprovals"`
	Budgets          BudgetInfo        `json:"budgets"`
}

// SidebarBadges holds badge counts for the sidebar.
type SidebarBadges struct {
	Inbox       int `json:"inbox"`
	Approvals   int `json:"approvals"`
	FailedRuns  int `json:"failedRuns"`
	JoinRequests int `json:"joinRequests"`
}

// GetDashboard returns a full dashboard summary for a company.
func (s *DashboardService) GetDashboard(ctx context.Context, companyUUID string) (*DashboardSummary, error) {
	// Verify company exists and get budget info.
	var companyInfo struct {
		BudgetMonthlyCents int `db:"budget_monthly_cents"`
	}
	err := s.db.GetContext(ctx, &companyInfo, `
		SELECT budget_monthly_cents FROM companies WHERE uuid = $1
	`, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Company not found")
	}
	if err != nil {
		return nil, err
	}

	// Agent status counts.
	type agentRow struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	var agentRows []agentRow
	if err := s.db.SelectContext(ctx, &agentRows, `
		SELECT status, COUNT(*)::int AS count
		FROM agents
		WHERE company_uuid = $1
		GROUP BY status
	`, companyUUID); err != nil {
		return nil, err
	}

	agentCounts := AgentStatusCounts{}
	for _, row := range agentRows {
		switch row.Status {
		case "idle":
			agentCounts.Active += row.Count
		case "running":
			agentCounts.Running += row.Count
		case "paused":
			agentCounts.Paused += row.Count
		case "error":
			agentCounts.Error += row.Count
		}
	}

	// Issue status counts.
	type issueRow struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	var issueRows []issueRow
	if err := s.db.SelectContext(ctx, &issueRows, `
		SELECT status, COUNT(*)::int AS count
		FROM issues
		WHERE company_uuid = $1
		GROUP BY status
	`, companyUUID); err != nil {
		return nil, err
	}

	issueCounts := IssueCounts{}
	for _, row := range issueRows {
		switch row.Status {
		case "in_progress":
			issueCounts.InProgress += row.Count
			issueCounts.Open += row.Count
		case "blocked":
			issueCounts.Blocked += row.Count
			issueCounts.Open += row.Count
		case "done":
			issueCounts.Done += row.Count
		case "cancelled":
			// excluded from open
		default:
			issueCounts.Open += row.Count
		}
	}

	// Monthly cost spend.
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var monthSpendCents int
	if err := s.db.GetContext(ctx, &monthSpendCents, `
		SELECT COALESCE(SUM(cost_cents), 0)::int
		FROM cost_events
		WHERE company_uuid = $1
		  AND occurred_at >= $2
	`, companyUUID, monthStart); err != nil {
		return nil, err
	}

	utilization := 0.0
	if companyInfo.BudgetMonthlyCents > 0 {
		utilization = float64(monthSpendCents) / float64(companyInfo.BudgetMonthlyCents) * 100
		// Round to 2 decimal places.
		utilization = float64(int(utilization*100)) / 100
	}

	// Pending approvals.
	var pendingApprovals int
	if err := s.db.GetContext(ctx, &pendingApprovals, `
		SELECT COUNT(*)::int FROM approvals
		WHERE company_uuid = $1 AND status = 'pending'
	`, companyUUID); err != nil {
		return nil, err
	}

	// Budget incidents.
	var activeIncidents int
	if err := s.db.GetContext(ctx, &activeIncidents, `
		SELECT COUNT(*)::int FROM budget_incidents
		WHERE company_uuid = $1 AND status = 'active'
	`, companyUUID); err != nil {
		return nil, err
	}

	var budgetPendingApprovals int
	if err := s.db.GetContext(ctx, &budgetPendingApprovals, `
		SELECT COUNT(*)::int FROM budget_incidents
		WHERE company_uuid = $1 AND status = 'pending_approval'
	`, companyUUID); err != nil {
		return nil, err
	}

	var pausedAgents int
	if err := s.db.GetContext(ctx, &pausedAgents, `
		SELECT COUNT(*)::int FROM agents
		WHERE company_uuid = $1 AND status = 'paused'
	`, companyUUID); err != nil {
		return nil, err
	}

	return &DashboardSummary{
		CompanyUUID: companyUUID,
		Agents:      agentCounts,
		Tasks:       issueCounts,
		Costs: CostInfo{
			MonthSpendCents:         monthSpendCents,
			MonthBudgetCents:        companyInfo.BudgetMonthlyCents,
			MonthUtilizationPercent: utilization,
		},
		PendingApprovals: pendingApprovals,
		Budgets: BudgetInfo{
			ActiveIncidents:  activeIncidents,
			PendingApprovals: budgetPendingApprovals,
			PausedAgents:     pausedAgents,
			PausedProjects:   0,
		},
	}, nil
}

// GetSidebarBadges returns badge counts for the sidebar navigation.
func (s *DashboardService) GetSidebarBadges(ctx context.Context, companyUUID string) (*SidebarBadges, error) {
	// Actionable approvals.
	var approvalsCount int
	if err := s.db.GetContext(ctx, &approvalsCount, `
		SELECT COUNT(*)::int FROM approvals
		WHERE company_uuid = $1 AND status IN ('pending', 'revision_requested')
	`, companyUUID); err != nil {
		return nil, err
	}

	// Failed runs: latest run per non-terminated agent.
	var failedRuns int
	if err := s.db.GetContext(ctx, &failedRuns, `
		SELECT COUNT(*)::int
		FROM (
			SELECT DISTINCT ON (hr.agent_uuid) hr.status
			FROM heartbeat_runs hr
			JOIN agents a ON a.uuid = hr.agent_uuid
			WHERE hr.company_uuid = $1
			  AND a.company_uuid  = $1
			  AND a.status        != 'terminated'
			ORDER BY hr.agent_uuid, hr.created_at DESC
		) latest
		WHERE latest.status IN ('failed', 'timed_out')
	`, companyUUID); err != nil {
		return nil, err
	}

	badges := &SidebarBadges{
		Approvals:  approvalsCount,
		FailedRuns: failedRuns,
		Inbox:      approvalsCount + failedRuns,
	}
	return badges, nil
}
