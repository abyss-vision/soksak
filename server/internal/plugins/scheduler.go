package plugins

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

// JobHandler is called when a scheduled job fires. pluginUUID and jobKey
// identify which job is executing.
type JobHandler func(ctx context.Context, pluginUUID, jobKey string) error

// ScheduledJob defines the registration parameters for a plugin job.
type ScheduledJob struct {
	// PluginUUID is the UUID of the owning plugin row in the plugins table.
	PluginUUID string
	// JobKey is the unique job identifier within the plugin.
	JobKey string
	// Schedule is a standard cron expression (5 or 6 fields).
	Schedule string
	// Handler is invoked each time the cron fires.
	Handler JobHandler
}

// jobEntry is an internal runtime entry for a registered job.
type jobEntry struct {
	job    ScheduledJob
	cronID cron.EntryID
}

// JobScheduler schedules cron jobs for plugins and persists execution records
// to the plugin_jobs and plugin_job_runs tables.
type JobScheduler struct {
	db   *sqlx.DB
	cr   *cron.Cron
	mu   sync.Mutex
	jobs map[string]*jobEntry // key: pluginUUID+"/"+jobKey
}

// NewJobScheduler creates a JobScheduler. Call Start() to begin scheduling.
func NewJobScheduler(db *sqlx.DB) *JobScheduler {
	return &JobScheduler{
		db:   db,
		cr:   cron.New(cron.WithSeconds()),
		jobs: make(map[string]*jobEntry),
	}
}

// Start begins the background cron ticker.
func (s *JobScheduler) Start() {
	s.cr.Start()
}

// Stop halts the cron ticker and waits for any in-flight jobs to finish.
func (s *JobScheduler) Stop() {
	ctx := s.cr.Stop()
	<-ctx.Done()
}

func jobKey(pluginUUID, jobKey string) string {
	return pluginUUID + "/" + jobKey
}

// Schedule registers a new cron job. Calling Schedule with the same
// pluginUUID+jobKey replaces any previously registered entry.
func (s *JobScheduler) Schedule(job ScheduledJob) error {
	if err := validateCron(job.Schedule); err != nil {
		return fmt.Errorf("invalid cron expression %q for job %s/%s: %w",
			job.Schedule, job.PluginUUID, job.JobKey, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	k := jobKey(job.PluginUUID, job.JobKey)

	// Remove existing entry if present.
	if existing, ok := s.jobs[k]; ok {
		s.cr.Remove(existing.cronID)
		delete(s.jobs, k)
	}

	id, err := s.cr.AddFunc(job.Schedule, s.buildRunner(job))
	if err != nil {
		return fmt.Errorf("schedule job %s: %w", k, err)
	}

	s.jobs[k] = &jobEntry{job: job, cronID: id}
	return nil
}

// Cancel removes the cron job identified by (pluginUUID, key).
// Returns an error if the job is not currently scheduled.
func (s *JobScheduler) Cancel(pluginUUID, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := jobKey(pluginUUID, key)
	entry, ok := s.jobs[k]
	if !ok {
		return fmt.Errorf("job not scheduled: %s", k)
	}
	s.cr.Remove(entry.cronID)
	delete(s.jobs, k)
	return nil
}

// ListJobs returns all currently scheduled jobs.
func (s *JobScheduler) ListJobs() []ScheduledJob {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobs := make([]ScheduledJob, 0, len(s.jobs))
	for _, e := range s.jobs {
		jobs = append(jobs, e.job)
	}
	return jobs
}

// RunPending is a manual trigger that executes all jobs whose next_run_at is
// in the past. Useful for bootstrapping after a server restart. Each job
// executes in its own goroutine; errors are logged but do not halt others.
func (s *JobScheduler) RunPending(ctx context.Context) {
	s.mu.Lock()
	snapshot := make([]*jobEntry, 0, len(s.jobs))
	for _, e := range s.jobs {
		snapshot = append(snapshot, e)
	}
	s.mu.Unlock()

	now := time.Now()
	for _, e := range snapshot {
		nextRunAt := s.nextRunAt(ctx, e.job.PluginUUID, e.job.JobKey)
		if nextRunAt != nil && nextRunAt.Before(now) {
			go s.buildRunner(e.job)()
		}
	}
}

// nextRunAt queries next_run_at from the plugin_jobs table.
func (s *JobScheduler) nextRunAt(ctx context.Context, pluginUUID, key string) *time.Time {
	var t *time.Time
	const q = `SELECT next_run_at FROM plugin_jobs WHERE plugin_uuid = $1 AND job_key = $2 LIMIT 1`
	_ = s.db.QueryRowContext(ctx, q, pluginUUID, key).Scan(&t)
	return t
}

// buildRunner returns the cron function for job. It records a run row and
// calls the user-supplied handler.
func (s *JobScheduler) buildRunner(job ScheduledJob) func() {
	return func() {
		ctx := context.Background()

		runUUID := uuid.NewString()
		startedAt := time.Now()

		// Resolve plugin_jobs.uuid for the FK.
		var jobUUID string
		const selectJobUUID = `SELECT uuid FROM plugin_jobs WHERE plugin_uuid = $1 AND job_key = $2 LIMIT 1`
		err := s.db.QueryRowContext(ctx, selectJobUUID, job.PluginUUID, job.JobKey).Scan(&jobUUID)
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("job_scheduler: plugin_jobs row missing, skipping run",
				"plugin_uuid", job.PluginUUID,
				"job_key", job.JobKey)
			return
		}
		if err != nil {
			slog.Error("job_scheduler: lookup plugin_jobs uuid", "error", err)
			return
		}

		// Insert a run record.
		const insertRun = `
			INSERT INTO plugin_job_runs
				(uuid, job_uuid, plugin_uuid, trigger, status, started_at, created_at)
			VALUES ($1, $2, $3, 'schedule', 'running', $4, now())`
		if _, err := s.db.ExecContext(ctx, insertRun, runUUID, jobUUID, job.PluginUUID, startedAt); err != nil {
			slog.Error("job_scheduler: insert job run", "error", err)
			return
		}

		// Invoke the handler.
		handlerErr := job.Handler(ctx, job.PluginUUID, job.JobKey)
		finishedAt := time.Now()
		durationMs := int(finishedAt.Sub(startedAt).Milliseconds())

		// Update run record.
		status := "succeeded"
		var errMsg *string
		if handlerErr != nil {
			status = "failed"
			msg := handlerErr.Error()
			errMsg = &msg
		}

		const updateRun = `
			UPDATE plugin_job_runs
			SET status = $1, duration_ms = $2, error = $3, finished_at = $4
			WHERE uuid = $5`
		if _, err := s.db.ExecContext(ctx, updateRun, status, durationMs, errMsg, finishedAt, runUUID); err != nil {
			slog.Error("job_scheduler: update job run", "error", err)
		}

		// Advance next_run_at.
		s.advanceNextRun(ctx, jobUUID, job.Schedule, finishedAt)
	}
}

// advanceNextRun computes the next cron tick and updates the plugin_jobs row.
func (s *JobScheduler) advanceNextRun(ctx context.Context, jobUUID, schedule string, from time.Time) {
	next, err := nextCronTick(schedule, from)
	if err != nil {
		return
	}
	const q = `UPDATE plugin_jobs SET last_run_at = $1, next_run_at = $2, updated_at = now() WHERE uuid = $3`
	if _, err := s.db.ExecContext(ctx, q, from, next, jobUUID); err != nil {
		slog.Error("job_scheduler: advance next_run_at", "error", err)
	}
}

// validateCron returns a non-nil error if schedule is not a valid cron expression.
func validateCron(schedule string) error {
	p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := p.Parse(schedule)
	if err != nil {
		// Also try standard 5-field cron.
		p2 := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		_, err2 := p2.Parse(schedule)
		if err2 != nil {
			return err
		}
	}
	return nil
}

// nextCronTick returns the next activation time after from for schedule.
func nextCronTick(schedule string, from time.Time) (time.Time, error) {
	p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	s, err := p.Parse(schedule)
	if err != nil {
		// Fall back to 5-field parser.
		p2 := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		s2, err2 := p2.Parse(schedule)
		if err2 != nil {
			return time.Time{}, fmt.Errorf("parse cron %q: %w", schedule, err)
		}
		s = s2
	}
	return s.Next(from), nil
}
