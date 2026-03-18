package plugins

import (
	"context"
	"testing"
	"time"
)

func TestJobKey(t *testing.T) {
	got := jobKey("plugin-uuid-1", "cleanup")
	want := "plugin-uuid-1/cleanup"
	if got != want {
		t.Errorf("jobKey = %q, want %q", got, want)
	}
}

func TestValidateCron_Valid(t *testing.T) {
	cases := []string{
		"0 * * * *",          // every hour (5-field)
		"0 0 * * *",          // midnight daily
		"*/5 * * * *",        // every 5 minutes
		"0 0 0 * * *",        // every midnight (6-field with seconds)
		"*/30 * * * * *",     // every 30 seconds (6-field)
	}
	for _, c := range cases {
		if err := validateCron(c); err != nil {
			t.Errorf("validateCron(%q): unexpected error: %v", c, err)
		}
	}
}

func TestValidateCron_Invalid(t *testing.T) {
	cases := []string{
		"",
		"not a cron",
		"99 99 99 99 99",
	}
	for _, c := range cases {
		if err := validateCron(c); err == nil {
			t.Errorf("validateCron(%q): expected error, got nil", c)
		}
	}
}

func TestNextCronTick(t *testing.T) {
	// 5-field: every minute
	from := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	next, err := nextCronTick("* * * * *", from)
	if err != nil {
		t.Fatalf("nextCronTick: unexpected error: %v", err)
	}
	if !next.After(from) {
		t.Errorf("next tick %v is not after from %v", next, from)
	}

	// 6-field: every second
	next2, err := nextCronTick("* * * * * *", from)
	if err != nil {
		t.Fatalf("nextCronTick (6-field): unexpected error: %v", err)
	}
	if !next2.After(from) {
		t.Errorf("next tick %v is not after from %v", next2, from)
	}
}

func TestNextCronTick_Invalid(t *testing.T) {
	_, err := nextCronTick("not valid cron", time.Now())
	if err == nil {
		t.Fatal("nextCronTick invalid: expected error, got nil")
	}
}

func TestJobScheduler_ScheduleAndList(t *testing.T) {
	// NewJobScheduler requires a db but we only test Schedule+Cancel+List
	// which use the in-memory map and the cron library (no DB calls in those paths).
	sched := NewJobScheduler(nil)

	job := ScheduledJob{
		PluginUUID: "plug-1",
		JobKey:     "sweep",
		Schedule:   "0 0 * * * *", // 6-field: every hour at second 0
		Handler: func(_ context.Context, _, _ string) error {
			return nil
		},
	}

	if err := sched.Schedule(job); err != nil {
		t.Fatalf("Schedule: unexpected error: %v", err)
	}

	jobs := sched.ListJobs()
	if len(jobs) != 1 {
		t.Fatalf("ListJobs: expected 1, got %d", len(jobs))
	}
	if jobs[0].JobKey != "sweep" {
		t.Errorf("JobKey = %q, want %q", jobs[0].JobKey, "sweep")
	}
}

func TestJobScheduler_Cancel(t *testing.T) {
	sched := NewJobScheduler(nil)

	job := ScheduledJob{
		PluginUUID: "plug-2",
		JobKey:     "audit",
		Schedule:   "0 0 0 * * *", // 6-field: midnight
		Handler:    func(_ context.Context, _, _ string) error { return nil },
	}
	_ = sched.Schedule(job)

	if err := sched.Cancel("plug-2", "audit"); err != nil {
		t.Fatalf("Cancel: unexpected error: %v", err)
	}

	jobs := sched.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("ListJobs after Cancel: expected 0, got %d", len(jobs))
	}
}

func TestJobScheduler_Cancel_Missing(t *testing.T) {
	sched := NewJobScheduler(nil)
	err := sched.Cancel("nonexistent", "job")
	if err == nil {
		t.Fatal("Cancel missing: expected error, got nil")
	}
}

func TestJobScheduler_Schedule_InvalidCron(t *testing.T) {
	sched := NewJobScheduler(nil)
	job := ScheduledJob{
		PluginUUID: "plug-3",
		JobKey:     "bad",
		Schedule:   "not a cron",
		Handler:    func(_ context.Context, _, _ string) error { return nil },
	}
	if err := sched.Schedule(job); err == nil {
		t.Fatal("Schedule invalid cron: expected error, got nil")
	}
}

func TestJobScheduler_Schedule_Replace(t *testing.T) {
	sched := NewJobScheduler(nil)

	job1 := ScheduledJob{
		PluginUUID: "plug-4",
		JobKey:     "report",
		Schedule:   "0 0 * * * *", // 6-field: every hour
		Handler:    func(_ context.Context, _, _ string) error { return nil },
	}
	job2 := ScheduledJob{
		PluginUUID: "plug-4",
		JobKey:     "report",
		Schedule:   "0 0 0 * * *", // 6-field: midnight (different schedule)
		Handler:    func(_ context.Context, _, _ string) error { return nil },
	}

	_ = sched.Schedule(job1)
	_ = sched.Schedule(job2) // should replace job1

	jobs := sched.ListJobs()
	if len(jobs) != 1 {
		t.Fatalf("after replace: expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Schedule != "0 0 0 * * *" {
		t.Errorf("Schedule = %q, want %q", jobs[0].Schedule, "0 0 0 * * *")
	}
}
