//go:build windows

package runner

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

var notifySignals = []os.Signal{os.Interrupt}

// shell runs command through cmd.exe. CmdLine is set directly so Go does not
// re-quote the command.
func shell(command string) *exec.Cmd {
	c := exec.Command("cmd.exe")
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: `cmd.exe /d /s /c "` + command + `"`}
	return c
}

func setGroup(*exec.Cmd) {}

// The console delivers Ctrl+C to every attached process.
func interruptTree(*exec.Cmd) {}

func terminateTree(c *exec.Cmd) {
	if c != nil && c.Process != nil {
		_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(c.Process.Pid)).Run()
	}
}

func killTree(c *exec.Cmd) { terminateTree(c) }

func forwardToChild(*exec.Cmd, os.Signal) {}
