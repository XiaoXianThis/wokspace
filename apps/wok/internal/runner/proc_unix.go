//go:build !windows

package runner

import (
	"os"
	"os/exec"
	"syscall"
)

var notifySignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGHUP}

func shell(command string) *exec.Cmd {
	return exec.Command("sh", "-c", command)
}

// setGroup puts the child in its own process group so that its whole tree
// can be signalled.
func setGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func signalGroup(c *exec.Cmd, s syscall.Signal) {
	if c != nil && c.Process != nil {
		_ = syscall.Kill(-c.Process.Pid, s)
	}
}

func interruptTree(c *exec.Cmd) { signalGroup(c, syscall.SIGINT) }
func terminateTree(c *exec.Cmd) { signalGroup(c, syscall.SIGTERM) }
func killTree(c *exec.Cmd)      { signalGroup(c, syscall.SIGKILL) }

// forwardToChild passes signals other than Ctrl+C, which the terminal
// already delivers to the child's process group.
func forwardToChild(c *exec.Cmd, s os.Signal) {
	if s != os.Interrupt && c.Process != nil {
		_ = c.Process.Signal(s)
	}
}
