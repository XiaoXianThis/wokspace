// Package scaffold writes the wokspace skeleton and new apps.
package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DiskName maps a template path to its path on disk: a segment named
// dot-x becomes .x. Template dotfiles are stored this way so that they do not
// take effect inside the wok source tree.
func DiskName(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, "dot-") {
			segs[i] = "." + strings.TrimPrefix(s, "dot-")
		}
	}
	return strings.Join(segs, "/")
}

// SkeletonDirs lists template directories that contain a README.md, as
// slash paths. These are the frozen skeleton.
func SkeletonDirs(tmpl fs.FS) []string {
	var dirs []string
	_ = fs.WalkDir(tmpl, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && path.Base(p) == "README.md" && path.Dir(p) != "." {
			dirs = append(dirs, path.Dir(p))
		}
		return nil
	})
	sort.Strings(dirs)
	return dirs
}

// InitOptions controls Init.
type InitOptions struct {
	Name     string // project name for docs/README.md; defaults to the directory name
	Existing bool   // allow a non-empty directory; existing files are kept
	Git      bool   // run git init when the directory is not a repository
}

// InitResult reports what Init did.
type InitResult struct {
	Dir     string
	Created []string
	Skipped []string
	GitInit bool
	GitErr  error
}

var ignorableAtInit = map[string]bool{".git": true, ".DS_Store": true, "Thumbs.db": true, "desktop.ini": true}

// Init writes the skeleton into dir. It never overwrites files.
func Init(tmpl fs.FS, dir string, opt InitOptions) (*InitResult, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	if !opt.Existing {
		entries, err := os.ReadDir(abs)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !ignorableAtInit[e.Name()] {
				return nil, fmt.Errorf("%s is not empty; use --existing to add the missing skeleton files without overwriting anything", abs)
			}
		}
	}
	name := opt.Name
	if name == "" {
		name = filepath.Base(abs)
	}
	res := &InitResult{Dir: abs}
	err = fs.WalkDir(tmpl, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := DiskName(p)
		dest := filepath.Join(abs, filepath.FromSlash(rel))
		if _, err := os.Lstat(dest); err == nil {
			res.Skipped = append(res.Skipped, rel)
			return nil
		}
		data, err := fs.ReadFile(tmpl, p)
		if err != nil {
			return err
		}
		if p == "docs/README.md" {
			data = []byte(strings.ReplaceAll(string(data), "{{name}}", name))
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return err
		}
		res.Created = append(res.Created, rel)
		return nil
	})
	if err != nil {
		return res, err
	}
	if opt.Git {
		if _, err := os.Stat(filepath.Join(abs, ".git")); os.IsNotExist(err) {
			if _, err := exec.LookPath("git"); err == nil {
				c := exec.Command("git", "init", "-q")
				c.Dir = abs
				if out, err := c.CombinedOutput(); err != nil {
					res.GitErr = fmt.Errorf("git init: %v: %s", err, strings.TrimSpace(string(out)))
				} else {
					res.GitInit = true
				}
			}
		}
	}
	return res, nil
}

var appName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var reserved = map[string]bool{"con": true, "prn": true, "aux": true, "nul": true}

func init() {
	for i := 1; i <= 9; i++ {
		reserved[fmt.Sprintf("com%d", i)] = true
		reserved[fmt.Sprintf("lpt%d", i)] = true
	}
}

// IsReserved reports whether name (or its part before the first dot) is a
// Windows reserved device name.
func IsReserved(name string) bool {
	stem := strings.ToLower(name)
	if i := strings.IndexByte(stem, '.'); i >= 0 {
		stem = stem[:i]
	}
	return reserved[strings.TrimRight(stem, " ")]
}

// ValidateAppName checks an app name against the naming rules.
func ValidateAppName(name string) error {
	if !appName.MatchString(name) {
		return fmt.Errorf("app name %q must be lowercase kebab-case, such as web or api-server", name)
	}
	if IsReserved(name) {
		return fmt.Errorf("app name %q is a reserved name on Windows", name)
	}
	return nil
}

// NewAppResult reports what NewApp did.
type NewAppResult struct {
	AgentsPath string
	Registered bool
}

const templateHeading = "App AGENTS.md template"

// NewApp creates apps/<name>/AGENTS.md from the template block in the
// project's apps/README.md (falling back to the built-in template) and adds
// the app to the list in apps/README.md.
func NewApp(root string, tmpl fs.FS, name string) (*NewAppResult, error) {
	if err := ValidateAppName(name); err != nil {
		return nil, err
	}
	appDir := filepath.Join(root, "apps", name)
	// An existing directory (for example from `cargo new`) is adopted, as
	// long as it has no AGENTS.md yet.
	if _, err := os.Stat(filepath.Join(appDir, "AGENTS.md")); err == nil {
		return nil, fmt.Errorf("apps/%s/AGENTS.md already exists", name)
	}
	readmePath := filepath.Join(root, "apps", "README.md")
	readme, readErr := os.ReadFile(readmePath)

	body, ok := "", false
	if readErr == nil {
		body, ok = templateBlock(string(readme))
	}
	if !ok {
		builtin, err := fs.ReadFile(tmpl, "apps/README.md")
		if err != nil {
			return nil, err
		}
		if body, ok = templateBlock(string(builtin)); !ok {
			return nil, errors.New("built-in app template is missing")
		}
	}
	body = strings.ReplaceAll(body, "<name>", name)

	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return nil, err
	}
	agents := filepath.Join(appDir, "AGENTS.md")
	if err := os.WriteFile(agents, []byte(body), 0o644); err != nil {
		return nil, err
	}
	res := &NewAppResult{AgentsPath: agents}
	if readErr == nil {
		if updated, ok := register(string(readme), name); ok {
			if err := os.WriteFile(readmePath, []byte(updated), 0o644); err != nil {
				return res, err
			}
			res.Registered = true
		}
	}
	return res, nil
}

func splitLines(s string) ([]string, string) {
	nl := "\n"
	if strings.Contains(s, "\r\n") {
		nl = "\r\n"
	}
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n"), nl
}

func heading(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "## ") {
		return "", false
	}
	return strings.TrimSpace(t[3:]), true
}

// templateBlock returns the first fenced block under the template heading.
func templateBlock(readme string) (string, bool) {
	lines, _ := splitLines(readme)
	in := false
	for i, l := range lines {
		if h, ok := heading(l); ok {
			in = strings.EqualFold(h, templateHeading)
			continue
		}
		t := strings.TrimSpace(l)
		if !in || !strings.HasPrefix(t, "```") {
			continue
		}
		fence := t[:len(t)-len(strings.TrimLeft(t, "`"))]
		var body []string
		for _, l2 := range lines[i+1:] {
			t2 := strings.TrimSpace(l2)
			if strings.HasPrefix(t2, fence) && strings.Trim(t2, "`") == "" {
				return strings.Join(body, "\n") + "\n", true
			}
			body = append(body, l2)
		}
		return "", false
	}
	return "", false
}

// register appends a row for name to the first table under "## Apps".
func register(readme, name string) (string, bool) {
	lines, nl := splitLines(readme)
	in := false
	for i := 0; i < len(lines); i++ {
		if h, ok := heading(lines[i]); ok {
			in = strings.EqualFold(h, "Apps")
			continue
		}
		if !in || !strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
			continue
		}
		cols := strings.Count(strings.Trim(strings.TrimSpace(lines[i]), "|"), "|") + 1
		end := i
		for end < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[end]), "|") {
			end++
		}
		row := "| " + name + " |" + strings.Repeat("  |", cols-1)
		out := append([]string{}, lines[:end]...)
		out = append(out, row)
		out = append(out, lines[end:]...)
		return strings.Join(out, nl), true
	}
	return "", false
}
