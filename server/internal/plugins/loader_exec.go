package plugins

import (
	"os/exec"
)

// commandForPath returns an *exec.Cmd that starts the plugin binary at path.
func commandForPath(path string) *exec.Cmd {
	return exec.Command(path) //nolint:gosec
}
