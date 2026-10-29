// Package runner runs task commands through the platform shell, one at a time
// with the terminal attached, or several in parallel with prefixed output.
package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/XiaoXianThis/wokspace/apps/wok/internal/ui"
)

// Job is one command to run.
type Job struct {
	Label   string
	Dir     string
	Command string
	Env     []string
}

func (j Job) cmd() *exec.Cmd {
	c := shell(j.Command)
	c.Dir = j.Dir
	c.Env = j.Env
	return c
}

// Run runs a job with the terminal attached and returns its exit code.
func Run(j Job) int {
	c := j.cmd()
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr

	// The terminal delivers Ctrl+C to the child directly; wok stays alive
	// to report the child's exit code.
	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, notifySignals...)
	defer signal.Stop(sigs)

	if err := c.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", ui.Red("wok:"), err)
		return 127
	}
	done := make(chan struct{})
	go func() {
		for {
			select {
			case s := <-sigs:
				forwardToChild(c, s)
			case <-done:
				return
			}
		}
	}()
	err := c.Wait()
	close(done)
	return exitCode(err)
}

// RunParallel runs jobs at once, prefixing each output line with the job's
// label. When one job fails, the others are stopped. Returns the first
// non-zero exit code.
func RunParallel(jobs []Job, w io.Writer) int {
	width := 0
	for _, j := range jobs {
		if len(j.Label) > width {
			width = len(j.Label)
		}
	}
	var mu sync.Mutex
	writeLine := func(prefix, line string) {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprint(w, prefix, line)
	}

	type result struct{ i, code int }
	results := make(chan result, len(jobs))
	cmds := make([]*exec.Cmd, len(jobs))
	running := make([]bool, len(jobs))
	prefixes := make([]string, len(jobs))
	var readers sync.WaitGroup

	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, notifySignals...)
	defer signal.Stop(sigs)

	exit := 0
	stopOthers := func(except int) {
		for k, c := range cmds {
			if k != except && running[k] {
				terminateTree(c)
			}
		}
	}

	for i, j := range jobs {
		prefixes[i] = ui.Label(i, fmt.Sprintf("%-*s |", width, j.Label)) + " "
		c := j.cmd()
		pr, pw, err := os.Pipe()
		if err != nil {
			results <- result{i, 127}
			continue
		}
		c.Stdout, c.Stderr = pw, pw
		setGroup(c)
		if err := c.Start(); err != nil {
			pw.Close()
			pr.Close()
			writeLine(prefixes[i], ui.Red(err.Error())+"\n")
			results <- result{i, 127}
			continue
		}
		pw.Close()
		cmds[i], running[i] = c, true
		readers.Add(1)
		go func(prefix string, r *os.File) {
			defer readers.Done()
			defer r.Close()
			br := bufio.NewReader(r)
			for {
				line, err := br.ReadString('\n')
				if line != "" {
					if !strings.HasSuffix(line, "\n") {
						line += "\n"
					}
					writeLine(prefix, line)
				}
				if err != nil {
					return
				}
			}
		}(prefixes[i], pr)
		go func(i int, c *exec.Cmd) { results <- result{i, exitCode(c.Wait())} }(i, c)
	}

	interrupts := 0
	for remaining := len(jobs); remaining > 0; {
		select {
		case r := <-results:
			remaining--
			running[r.i] = false
			if r.code == 0 {
				writeLine(prefixes[r.i], ui.Dim("exited")+"\n")
				continue
			}
			writeLine(prefixes[r.i], ui.Red(fmt.Sprintf("exited with code %d", r.code))+"\n")
			if exit == 0 {
				exit = r.code
				if interrupts == 0 {
					stopOthers(r.i)
				}
			}
		case <-sigs:
			interrupts++
			for k, c := range cmds {
				if running[k] {
					if interrupts == 1 {
						interruptTree(c)
					} else {
						killTree(c)
					}
				}
			}
		}
	}

	// Grandchildren may hold the pipes open briefly after their parents exit.
	done := make(chan struct{})
	go func() { readers.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
	if exit == 0 && interrupts > 0 {
		exit = 130
	}
	return exit
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if code := ee.ExitCode(); code >= 0 {
			return code
		}
		return 130 // killed by a signal
	}
	return 1
}

// QuoteArgs joins extra arguments for appending to a shell command.
func QuoteArgs(args []string, goos string) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = quote(a, goos)
	}
	return strings.Join(parts, " ")
}

func quote(a, goos string) string {
	if a != "" && strings.IndexFunc(a, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("-_./:=@,+%", r))
	}) < 0 {
		return a
	}
	if goos == "windows" {
		return `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
	}
	return "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
}
