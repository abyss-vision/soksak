package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ProcessConfig holds parameters for starting a new agent process.
type ProcessConfig struct {
	// Command and arguments to execute.
	Command string
	Args    []string

	// Env holds additional environment variables (KEY=VALUE format).
	Env []string

	// WorkDir is the working directory for the subprocess.
	WorkDir string

	// Identifiers used for publishing events.
	AgentID   string
	CompanyID string

	// Optional pre-assigned run ID. If empty, a UUID is generated.
	RunID string
}

// ProcessInfo summarises a running process for external consumers.
type ProcessInfo struct {
	RunID     string
	AgentID   string
	CompanyID string
	StartedAt time.Time
}

// Publisher is a narrow interface so the Manager can emit events without
// importing the full realtime package (avoids import cycle).
type Publisher interface {
	PublishRunLog(companyID, runID, stream, line string)
	PublishRunCompleted(companyID, runID string, exitCode int)
}

// agentProcess tracks a single running subprocess.
type agentProcess struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	runID     string
	agentID   string
	companyID string
	startedAt time.Time
	done      chan struct{}
}

// Manager owns the lifecycle of agent subprocesses.
type Manager struct {
	mu        sync.RWMutex
	processes map[string]*agentProcess
}

// New creates an empty Manager.
func New() *Manager {
	return &Manager{
		processes: make(map[string]*agentProcess),
	}
}

// Start spawns a new subprocess according to cfg, wires up stdio pipes, and
// registers it in the internal map. It returns the assigned runID.
func (m *Manager) Start(ctx context.Context, cfg ProcessConfig, pub Publisher) (string, error) {
	if cfg.Command == "" {
		return "", fmt.Errorf("process: command is required")
	}

	runID := cfg.RunID
	if runID == "" {
		runID = uuid.NewString()
	}

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	if cfg.WorkDir != "" {
		cmd.Dir = cfg.WorkDir
	}
	if len(cfg.Env) > 0 {
		cmd.Env = append(os.Environ(), cfg.Env...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("process: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("process: stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("process: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("process: start: %w", err)
	}

	ap := &agentProcess{
		cmd:       cmd,
		stdin:     stdin,
		runID:     runID,
		agentID:   cfg.AgentID,
		companyID: cfg.CompanyID,
		startedAt: time.Now(),
		done:      make(chan struct{}),
	}

	m.mu.Lock()
	m.processes[runID] = ap
	m.mu.Unlock()

	slog.Info("process: started",
		"runID", runID,
		"agentID", cfg.AgentID,
		"cmd", cfg.Command,
	)

	go m.streamOutput(ap, stdout, stderr, pub)

	return runID, nil
}

// Stop sends SIGTERM, waits up to 5 seconds, then escalates to SIGKILL.
func (m *Manager) Stop(runID string) error {
	m.mu.RLock()
	ap, ok := m.processes[runID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process: run %s not found", runID)
	}

	if err := ap.cmd.Process.Signal(os.Interrupt); err != nil {
		slog.Warn("process: SIGTERM failed, sending SIGKILL", "runID", runID, "err", err)
		return ap.cmd.Process.Kill()
	}

	select {
	case <-ap.done:
		return nil
	case <-time.After(5 * time.Second):
		slog.Warn("process: did not exit after SIGTERM, sending SIGKILL", "runID", runID)
		return ap.cmd.Process.Kill()
	}
}

// Signal sends an arbitrary signal to the process identified by runID.
func (m *Manager) Signal(runID string, sig os.Signal) error {
	m.mu.RLock()
	ap, ok := m.processes[runID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process: run %s not found", runID)
	}
	return ap.cmd.Process.Signal(sig)
}

// WriteStdin writes data followed by a newline to the process's stdin pipe.
func (m *Manager) WriteStdin(runID string, data string) error {
	m.mu.RLock()
	ap, ok := m.processes[runID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process: run %s not found", runID)
	}

	_, err := fmt.Fprintf(ap.stdin, "%s\n", data)
	return err
}

// GetRunningProcesses returns a snapshot of all currently running processes.
func (m *Manager) GetRunningProcesses() []ProcessInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProcessInfo, 0, len(m.processes))
	for _, ap := range m.processes {
		infos = append(infos, ProcessInfo{
			RunID:     ap.runID,
			AgentID:   ap.agentID,
			CompanyID: ap.companyID,
			StartedAt: ap.startedAt,
		})
	}
	return infos
}

// streamOutput fans out stdout and stderr lines to the hub and waits for the
// process to exit, then cleans up and publishes run.completed.
func (m *Manager) streamOutput(ap *agentProcess, stdout, stderr io.Reader, pub Publisher) {
	var wg sync.WaitGroup

	scanStream := func(r io.Reader, stream string) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			pub.PublishRunLog(ap.companyID, ap.runID, stream, line)
		}
		if err := scanner.Err(); err != nil {
			slog.Debug("process: scanner error", "stream", stream, "runID", ap.runID, "err", err)
		}
	}

	wg.Add(2)
	go scanStream(stdout, "stdout")
	go scanStream(stderr, "stderr")

	wg.Wait()

	exitCode := 0
	if err := ap.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			slog.Error("process: wait error", "runID", ap.runID, "err", err)
		}
	}

	close(ap.done)

	m.mu.Lock()
	delete(m.processes, ap.runID)
	m.mu.Unlock()

	slog.Info("process: exited", "runID", ap.runID, "exitCode", exitCode)
	pub.PublishRunCompleted(ap.companyID, ap.runID, exitCode)
}
