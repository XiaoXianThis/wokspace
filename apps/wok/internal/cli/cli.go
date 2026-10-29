// Package cli implements the wok commands.
package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/XiaoXianThis/wokspace/apps/wok/internal/check"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/project"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/runner"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/scaffold"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/tasks"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/ui"
)

// Env holds what the commands need from the outside world.
type Env struct {
	Version  string
	Template fs.FS
	Stdout   io.Writer
	Stderr   io.Writer
}

const usage = `wok - run a wokspace: one repository, any stack, built for agents

Usage:
  wok init [dir] [--name NAME] [--existing] [--no-git]
                               create a wokspace (default: current directory)
  wok new <app>                create apps/<app>/ with its AGENTS.md
  wok dev [app...]             start apps; several run in parallel
  wok build [app...] [--all]   build apps into build/<app>/
  wok test [app...] [--all]    test apps
  wok install [app...] [--all] install app dependencies
  wok run <app> <task> [-- args]
                               run any task from an app's task table
  wok ls [app]                 list apps and their tasks
  wok check                    check the layout against AGENTS.md
  wok version                  print the version

Inside apps/<app>/, the app can be omitted. Arguments after -- are passed
to the task command (single app only).

Tasks come from the task table in apps/<app>/AGENTS.md. Empty commands fall
back to what wok detects from the app's project files.

Docs: https://wokspace.wxian.cc
`

// Main runs wok with args (without the program name) and returns the exit code.
func Main(args []string, env Env) int {
	if len(args) == 0 {
		fmt.Fprint(env.Stdout, usage)
		return 0
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "-h", "--help", "help":
		fmt.Fprint(env.Stdout, usage)
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(env.Stdout, "wok %s (layout %d, %s/%s)\n", env.Version, project.LayoutVersion, runtime.GOOS, runtime.GOARCH)
		return 0
	case "init":
		return cmdInit(rest, env)
	case "new":
		return cmdNew(rest, env)
	case "dev", "build", "test", "install":
		return cmdTask(cmd, rest, env)
	case "run":
		return cmdRun(rest, env)
	case "ls", "list":
		return cmdLs(rest, env)
	case "check":
		return cmdCheck(rest, env)
	}
	return fail(env, "unknown command %q; run wok help", cmd)
}

func fail(env Env, format string, a ...any) int {
	fmt.Fprintf(env.Stderr, "%s %s\n", ui.Red("wok:"), fmt.Sprintf(format, a...))
	return 1
}

func findProject(env Env) (*project.Project, int) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fail(env, "%v", err)
	}
	p, err := project.Find(cwd)
	if errors.Is(err, project.ErrNotFound) {
		return nil, fail(env, "not inside a wokspace (no AGENTS.md next to an apps/ directory here or above); create one with wok init")
	}
	if err != nil {
		return nil, fail(env, "%v", err)
	}
	return p, 0
}

// ---- init ----

func cmdInit(args []string, env Env) int {
	var opt scaffold.InitOptions
	opt.Git = true
	dir := ""
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--existing":
			opt.Existing = true
		case a == "--no-git":
			opt.Git = false
		case a == "--name":
			if i+1 >= len(args) {
				return fail(env, "--name needs a value")
			}
			i++
			opt.Name = args[i]
		case strings.HasPrefix(a, "--name="):
			opt.Name = strings.TrimPrefix(a, "--name=")
		case strings.HasPrefix(a, "-"):
			return fail(env, "unknown flag %s", a)
		case dir == "":
			dir = a
		default:
			return fail(env, "init takes at most one directory")
		}
	}
	if dir == "" {
		dir = "."
	}
	res, err := scaffold.Init(env.Template, dir, opt)
	if err != nil {
		return fail(env, "%v", err)
	}
	out := env.Stdout
	fmt.Fprintf(out, "%s wokspace in %s (%s)\n", ui.Green("Created"), res.Dir, plural(len(res.Created), "file"))
	if len(res.Skipped) > 0 {
		fmt.Fprintf(out, "%s %s:\n", ui.Yellow("Kept existing"), plural(len(res.Skipped), "file"))
		for _, s := range res.Skipped {
			fmt.Fprintf(out, "  %s\n", s)
		}
	}
	switch {
	case res.GitInit:
		fmt.Fprintln(out, ui.Dim("Initialized a git repository."))
	case res.GitErr != nil:
		fmt.Fprintf(out, "%s %v\n", ui.Yellow("Skipped git:"), res.GitErr)
	}
	fmt.Fprintln(out, "\nNext:")
	if dir != "." {
		fmt.Fprintf(out, "  cd %s\n", dir)
	}
	fmt.Fprintln(out, "  wok new <app>      create your first app")
	fmt.Fprintln(out, "  wok check          check the layout")
	return 0
}

// ---- new ----

func cmdNew(args []string, env Env) int {
	if len(args) != 1 || strings.HasPrefix(args[0], "-") {
		return fail(env, "usage: wok new <app>")
	}
	p, code := findProject(env)
	if p == nil {
		return code
	}
	name := args[0]
	res, err := scaffold.NewApp(p.Root, env.Template, name)
	if err != nil {
		return fail(env, "%v", err)
	}
	out := env.Stdout
	fmt.Fprintf(out, "%s apps/%s/AGENTS.md\n", ui.Green("Created"), name)
	if res.Registered {
		fmt.Fprintf(out, "%s %s in apps/README.md\n", ui.Green("Listed"), name)
	} else {
		fmt.Fprintf(out, "%s no app list found in apps/README.md; add %s by hand\n", ui.Yellow("Note:"), name)
	}
	fmt.Fprintf(out, "\nNext:\n  1. initialize the app inside apps/%s/ with its own toolchain\n", name)
	fmt.Fprintf(out, "  2. fill in the task table in apps/%s/AGENTS.md (empty commands use detection)\n", name)
	fmt.Fprintf(out, "  3. wok ls %s       confirm the commands\n", name)
	return 0
}

// ---- tasks ----

type taskArgs struct {
	apps  []string
	all   bool
	extra []string
}

func parseTaskArgs(args []string) (taskArgs, error) {
	var t taskArgs
	for i, a := range args {
		switch {
		case a == "--":
			t.extra = args[i+1:]
			return t, nil
		case a == "--all" || a == "-a":
			t.all = true
		case strings.HasPrefix(a, "-"):
			return t, fmt.Errorf("unknown flag %s (pass task arguments after --)", a)
		default:
			t.apps = append(t.apps, a)
		}
	}
	return t, nil
}

func loadApp(p *project.Project, name string) (*tasks.App, error) {
	if !p.HasApp(name) {
		apps, _ := p.Apps()
		if len(apps) == 0 {
			return nil, fmt.Errorf("no app named %s; there are no apps yet (create one with wok new %s)", name, name)
		}
		return nil, fmt.Errorf("no app named %s; apps: %s", name, strings.Join(apps, ", "))
	}
	return tasks.Load(name, p.AppDir(name), runtime.GOOS)
}

func job(p *project.Project, app *tasks.App, t tasks.Task, extra []string) (runner.Job, error) {
	out := p.OutDir(app.Name)
	vars := map[string]string{"WOK_ROOT": p.Root, "WOK_APP": app.Name, "WOK_OUT": out, "WOK_TASK": t.Name}
	cmd := tasks.Expand(t.Command, vars)
	if len(extra) > 0 {
		cmd += " " + runner.QuoteArgs(extra, runtime.GOOS)
	}
	if t.Name == "build" {
		if err := os.MkdirAll(out, 0o755); err != nil {
			return runner.Job{}, err
		}
	}
	env := os.Environ()
	for k, v := range vars {
		env = append(env, k+"="+v)
	}
	return runner.Job{Label: app.Name, Dir: app.Dir, Command: cmd, Env: env}, nil
}

func missingTask(app *tasks.App, task string) error {
	return fmt.Errorf("%s has no %s command; add one to the task table in apps/%s/AGENTS.md", app.Name, task, app.Name)
}

func cmdTask(task string, args []string, env Env) int {
	ta, err := parseTaskArgs(args)
	if err != nil {
		return fail(env, "%v", err)
	}
	p, code := findProject(env)
	if p == nil {
		return code
	}
	names := ta.apps
	if ta.all {
		if len(names) > 0 {
			return fail(env, "give app names or --all, not both")
		}
		if names, err = p.Apps(); err != nil {
			return fail(env, "%v", err)
		}
		if len(names) == 0 {
			fmt.Fprintln(env.Stderr, "No apps yet. Create one with: wok new <app>")
			return 0
		}
	}
	if len(names) == 0 {
		cwd, _ := os.Getwd()
		if here := p.AppAt(cwd); here != "" {
			names = []string{here}
		} else {
			apps, _ := p.Apps()
			hint := "create one with wok new <app>"
			if len(apps) > 0 {
				hint = "apps: " + strings.Join(apps, ", ")
			}
			return fail(env, "which app? usage: wok %s <app>; %s", task, hint)
		}
	}
	if len(ta.extra) > 0 && len(names) != 1 {
		return fail(env, "arguments after -- need exactly one app")
	}

	// Resolve everything before running anything.
	var jobs []runner.Job
	var shown []string
	for _, name := range names {
		app, err := loadApp(p, name)
		if err != nil {
			return fail(env, "%v", err)
		}
		t, ok := app.Get(task)
		if !ok || !t.Defined() {
			if ta.all {
				fmt.Fprintf(env.Stderr, "%s %s has no %s command\n", ui.Dim("skip"), name, task)
				continue
			}
			return fail(env, "%v", missingTask(app, task))
		}
		j, err := job(p, app, t, ta.extra)
		if err != nil {
			return fail(env, "%v", err)
		}
		jobs = append(jobs, j)
		shown = append(shown, j.Command)
	}
	if len(jobs) == 0 {
		return 0
	}

	if task == "dev" && len(jobs) > 1 {
		for i, j := range jobs {
			fmt.Fprintf(env.Stderr, "%s %s\n", ui.Label(i, j.Label), ui.Dim(shown[i]))
			if ui.Enabled {
				jobs[i].Env = append(jobs[i].Env, "FORCE_COLOR=1", "CLICOLOR_FORCE=1")
			}
		}
		return runner.RunParallel(jobs, env.Stdout)
	}
	for i, j := range jobs {
		fmt.Fprintf(env.Stderr, "%s %s %s\n", ui.Cyan(j.Label), ui.Bold(task), ui.Dim(shown[i]))
		if code := runner.Run(j); code != 0 {
			fmt.Fprintf(env.Stderr, "%s %s %s failed with code %d\n", ui.Red("wok:"), j.Label, task, code)
			return code
		}
	}
	return 0
}

func cmdRun(args []string, env Env) int {
	var pos, extra []string
	for i, a := range args {
		if a == "--" {
			extra = args[i+1:]
			break
		}
		pos = append(pos, a)
	}
	if len(pos) != 2 {
		return fail(env, "usage: wok run <app> <task> [-- args]")
	}
	p, code := findProject(env)
	if p == nil {
		return code
	}
	app, err := loadApp(p, pos[0])
	if err != nil {
		return fail(env, "%v", err)
	}
	t, ok := app.Get(pos[1])
	if !ok || !t.Defined() {
		return fail(env, "%v", missingTask(app, pos[1]))
	}
	j, err := job(p, app, t, extra)
	if err != nil {
		return fail(env, "%v", err)
	}
	fmt.Fprintf(env.Stderr, "%s %s %s\n", ui.Cyan(app.Name), ui.Bold(t.Name), ui.Dim(j.Command))
	return runner.Run(j)
}

// ---- ls ----

func cmdLs(args []string, env Env) int {
	if len(args) > 1 {
		return fail(env, "usage: wok ls [app]")
	}
	p, code := findProject(env)
	if p == nil {
		return code
	}
	names := args
	if len(names) == 0 {
		var err error
		if names, err = p.Apps(); err != nil {
			return fail(env, "%v", err)
		}
	}
	out := env.Stdout
	if len(names) == 0 {
		fmt.Fprintln(out, "No apps yet. Create one with: wok new <app>")
		return 0
	}
	for i, name := range names {
		app, err := loadApp(p, name)
		if err != nil {
			return fail(env, "%v", err)
		}
		if i > 0 {
			fmt.Fprintln(out)
		}
		rel, _ := filepath.Rel(p.Root, app.Dir)
		fmt.Fprintf(out, "%s %s\n", ui.Bold(app.Name), ui.Dim(filepath.ToSlash(rel)))
		if !app.HasAgents {
			fmt.Fprintf(out, "  %s\n", ui.Yellow("no AGENTS.md"))
		}
		width, cmdWidth := 0, 0
		for _, t := range app.Tasks {
			width = max(width, len(t.Name))
			cmdWidth = max(cmdWidth, len(t.Command))
		}
		for _, t := range app.Tasks {
			name := fmt.Sprintf("%-*s", width, t.Name)
			if !t.Defined() {
				fmt.Fprintf(out, "  %s  %s\n", ui.Dim(name), ui.Dim("not set"))
				continue
			}
			fmt.Fprintf(out, "  %s  %-*s  %s\n", ui.Cyan(name), cmdWidth, t.Command, ui.Dim(t.Source))
		}
	}
	return 0
}

// ---- check ----

func cmdCheck(args []string, env Env) int {
	if len(args) > 0 {
		return fail(env, "usage: wok check")
	}
	p, code := findProject(env)
	if p == nil {
		return code
	}
	issues := check.Run(p, env.Template)
	errs, warns := 0, 0
	for _, is := range issues {
		label := ui.Yellow("warn ")
		if is.Severity == check.Error {
			label = ui.Red("error")
			errs++
		} else {
			warns++
		}
		fmt.Fprintf(env.Stdout, "%s %s  %s\n", label, ui.Bold(is.Path), is.Message)
	}
	switch {
	case len(issues) == 0:
		fmt.Fprintln(env.Stdout, ui.Green("Layout OK."))
	default:
		fmt.Fprintf(env.Stdout, "\n%s, %s\n", plural(errs, "error"), plural(warns, "warning"))
	}
	if errs > 0 {
		return 1
	}
	return 0
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
