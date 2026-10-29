package tasks

import (
	"os"
	"path/filepath"
	"strings"
)

// Standard tasks are always listed, even when nothing defines them.
var Standard = []string{"install", "dev", "test", "build"}

// SourceTable marks commands that come from the app's AGENTS.md.
const SourceTable = "AGENTS.md"

// Task is a resolved task. Command is empty when nothing defines it.
type Task struct {
	Name    string
	Command string
	Source  string
}

func (t Task) Defined() bool { return t.Command != "" }

// App is an app with its resolved tasks.
type App struct {
	Name      string
	Dir       string
	HasAgents bool
	HasTable  bool
	Tasks     []Task
	index     map[string]int
}

// Get returns the named task.
func (a *App) Get(name string) (Task, bool) {
	i, ok := a.index[strings.ToLower(name)]
	if !ok {
		return Task{}, false
	}
	return a.Tasks[i], true
}

// Load resolves the tasks of the app in dir for the given GOOS. Commands from
// the AGENTS.md task table win; empty table cells fall back to detection.
func Load(name, dir, goos string) (*App, error) {
	app := &App{Name: name, Dir: dir, index: map[string]int{}}
	set := func(t Task) {
		if i, ok := app.index[t.Name]; ok {
			app.Tasks[i] = t
			return
		}
		app.index[t.Name] = len(app.Tasks)
		app.Tasks = append(app.Tasks, t)
	}
	for _, s := range Standard {
		set(Task{Name: s})
	}

	var table Table
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	switch {
	case err == nil:
		app.HasAgents = true
		table = ParseTable(string(data))
		app.HasTable = table.Found
	case !os.IsNotExist(err):
		return nil, err
	}

	// Table rows define the order of non-standard tasks.
	for _, name := range table.Order {
		if _, ok := app.index[name]; !ok {
			set(Task{Name: name})
		}
	}
	for _, d := range detectAll(dir, goos) {
		for _, name := range d.order {
			if t, ok := app.Get(name); ok && t.Defined() {
				continue
			}
			set(Task{Name: name, Command: d.cmds[name], Source: "detected: " + d.source})
		}
	}
	for _, name := range table.Order {
		e := table.Entries[name]
		cmd := e.Command
		if goos == "windows" && e.Windows != "" {
			cmd = e.Windows
		}
		if cmd != "" {
			set(Task{Name: name, Command: cmd, Source: SourceTable})
		}
	}
	return app, nil
}

// Expand replaces $NAME and ${NAME} for the given variables only. Other
// variables are left for the shell.
func Expand(cmd string, vars map[string]string) string {
	var b strings.Builder
	for i := 0; i < len(cmd); {
		if cmd[i] == '$' {
			if strings.HasPrefix(cmd[i:], "${") {
				if end := strings.IndexByte(cmd[i:], '}'); end > 2 {
					if v, ok := vars[cmd[i+2:i+end]]; ok {
						b.WriteString(v)
						i += end + 1
						continue
					}
				}
			}
			j := i + 1
			for j < len(cmd) && isIdent(cmd[j]) {
				j++
			}
			if v, ok := vars[cmd[i+1:j]]; ok && j > i+1 {
				b.WriteString(v)
				i = j
				continue
			}
		}
		b.WriteByte(cmd[i])
		i++
	}
	return b.String()
}

func isIdent(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}
