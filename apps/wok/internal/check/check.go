// Package check verifies that a wokspace follows the layout rules in
// AGENTS.md.
package check

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/XiaoXianThis/wokspace/apps/wok/internal/project"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/scaffold"
	"github.com/XiaoXianThis/wokspace/apps/wok/internal/tasks"
)

// Severity of an issue.
type Severity int

const (
	Error Severity = iota
	Warning
)

// Issue is one finding.
type Issue struct {
	Severity Severity
	Path     string
	Message  string
}

type checker struct {
	p        *project.Project
	skeleton map[string]bool
	issues   []Issue
}

func (c *checker) errorf(p, format string, a ...any) {
	c.issues = append(c.issues, Issue{Error, p, fmt.Sprintf(format, a...)})
}

func (c *checker) warnf(p, format string, a ...any) {
	c.issues = append(c.issues, Issue{Warning, p, fmt.Sprintf(format, a...)})
}

var (
	required = []string{"AGENTS.md", ".gitignore", ".gitattributes", "apps", "build", "docs", "work"}
	optional = []string{".github", ".gitlab-ci.yml", "LICENSE", ".gitmodules"}
	local    = []string{
		".git", ".cursor", ".kiro", ".claude", ".codex", ".gemini", ".windsurf", ".continue",
		".vscode", ".idea", ".zed", ".fleet", "CLAUDE.md", "GEMINI.md",
	}
	junk     = map[string]bool{".DS_Store": true, "Thumbs.db": true, "desktop.ini": true}
	unstored = []string{"work/tmp", "work/cache", "work/logs"}
)

// Run checks the project. Issues are sorted errors first, then by path.
func Run(p *project.Project, tmpl fs.FS) []Issue {
	c := &checker{p: p, skeleton: map[string]bool{}}
	for _, d := range scaffold.SkeletonDirs(tmpl) {
		c.skeleton[d] = true
	}
	c.root()
	c.gitignore()
	c.agentsVersion()
	c.skeletonDirs()
	c.files()
	c.emptyDirs()
	c.apps()
	c.build()
	c.tmp()
	sort.SliceStable(c.issues, func(i, j int) bool {
		if c.issues[i].Severity != c.issues[j].Severity {
			return c.issues[i].Severity < c.issues[j].Severity
		}
		return c.issues[i].Path < c.issues[j].Path
	})
	return c.issues
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (c *checker) exists(slashPath string) bool {
	_, err := os.Stat(c.p.Path(slashPath))
	return err == nil
}

func (c *checker) root() {
	for _, name := range required {
		if !c.exists(name) {
			c.errorf(name, "required root entry is missing")
		}
	}
	if !c.exists(".git") {
		c.warnf(".git", "not a git repository; run git init")
	}
	entries, err := os.ReadDir(c.p.Root)
	if err != nil {
		c.errorf(".", "cannot read the root: %v", err)
		return
	}
	for _, e := range entries {
		n := e.Name()
		switch {
		case contains(required, n), contains(optional, n), contains(local, n), junk[n], strings.HasPrefix(n, ".aider"):
		default:
			c.errorf(n, "not allowed at the root (AGENTS.md section 1)")
		}
	}
}

func (c *checker) gitignore() {
	data, err := os.ReadFile(c.p.Path(".gitignore"))
	if err != nil {
		return
	}
	lines := map[string]bool{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		lines[strings.TrimSpace(sc.Text())] = true
	}
	if !lines["/*"] {
		c.errorf(".gitignore", "the root whitelist must start with /*")
	}
	whitelisted := func(name string) bool { return lines["!/"+name] || lines["!/"+name+"/"] }
	for _, name := range required {
		if !whitelisted(name) {
			c.errorf(".gitignore", "missing whitelist line !/%s", name)
		}
	}
	for _, name := range optional {
		if c.exists(name) && !whitelisted(name) {
			c.errorf(name, "optional root entry exists but .gitignore has no !/%s line, so git ignores it", name)
		}
	}
}

var layoutMarker = regexp.MustCompile(`<!--\s*wokspace-layout:\s*(\d+)\s*-->`)

func (c *checker) agentsVersion() {
	data, err := os.ReadFile(c.p.Path("AGENTS.md"))
	if err != nil {
		return
	}
	m := layoutMarker.FindSubmatch(data)
	if m == nil {
		c.warnf("AGENTS.md", "no <!-- wokspace-layout: N --> marker; wok cannot tell which layout version this is")
		return
	}
	v, _ := strconv.Atoi(string(m[1]))
	switch {
	case v < project.LayoutVersion:
		c.warnf("AGENTS.md", "layout version %d is older than wok's version %d; compare with the latest template", v, project.LayoutVersion)
	case v > project.LayoutVersion:
		c.warnf("AGENTS.md", "layout version %d is newer than this wok (%d); upgrade wok", v, project.LayoutVersion)
	}
}

// skeletonDirs checks that every skeleton directory and README exists and
// that no unknown directory sits next to skeleton directories.
func (c *checker) skeletonDirs() {
	parents := map[string]bool{}
	for d := range c.skeleton {
		if pd := path.Dir(d); pd != "." {
			parents[pd] = true
		}
		if !c.exists(d) {
			c.errorf(d, "skeleton directory is missing")
		} else if !c.exists(d + "/README.md") {
			c.errorf(d+"/README.md", "skeleton README is missing")
		}
	}
	for pd := range parents {
		entries, err := os.ReadDir(c.p.Path(pd))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() && !c.skeleton[pd+"/"+e.Name()] {
				c.errorf(pd+"/"+e.Name(), "not part of the skeleton (AGENTS.md section 2)")
			}
		}
	}
}

// listFiles returns files that git would commit (tracked or untracked but
// not ignored), as slash paths. Without git it walks the tree and skips
// what the template ignores.
func (c *checker) listFiles() []string {
	if c.exists(".git") {
		cmd := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
		cmd.Dir = c.p.Root
		if out, err := cmd.Output(); err == nil {
			var files []string
			for _, f := range strings.Split(string(out), "\x00") {
				if f != "" && c.exists(f) {
					files = append(files, f)
				}
			}
			return files
		}
	}
	var files []string
	skipDir := map[string]bool{"node_modules": true, "target": true, ".venv": true, "__pycache__": true}
	_ = filepath.WalkDir(c.p.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(c.p.Root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "." {
				return nil
			}
			if skipDir[d.Name()] || (!strings.Contains(rel, "/") && (contains(local, rel) || strings.HasPrefix(rel, ".aider"))) ||
				contains(local, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if junk[d.Name()] || under(rel, "build") || inUnstored(rel) {
			if !strings.HasSuffix(rel, "/README.md") || strings.Count(rel, "/") > 2 {
				return nil
			}
		}
		files = append(files, rel)
		return nil
	})
	return files
}

func under(p, dir string) bool { return strings.HasPrefix(p, dir+"/") }

func inUnstored(p string) bool {
	for _, d := range unstored {
		if under(p, d) {
			return true
		}
	}
	return false
}

var (
	kebab      = regexp.MustCompile(`^\.?[a-z0-9][a-z0-9._-]*$`)
	upperStems = map[string]bool{
		"README": true, "AGENTS": true, "CHANGELOG": true, "CONTRIBUTING": true,
		"SECURITY": true, "LICENSE": true, "CODE_OF_CONDUCT": true, "Dockerfile": true,
	}
	projectFiles = map[string]bool{
		"package.json": true, "Cargo.toml": true, "go.mod": true, "pyproject.toml": true, "pom.xml": true,
		"build.gradle": true, "build.gradle.kts": true, "settings.gradle": true, "settings.gradle.kts": true,
		"deno.json": true, "deno.jsonc": true, "composer.json": true, "Gemfile": true, "mix.exs": true,
		"CMakeLists.txt": true,
	}
	runnerFiles = map[string]bool{
		"Makefile": true, "makefile": true, "GNUmakefile": true, "justfile": true, "Justfile": true,
		".justfile": true, "Taskfile.yml": true, "Taskfile.yaml": true,
	}
)

const maxFileSize = 10 << 20

func (c *checker) files() {
	files := c.listFiles()
	seen := map[string]bool{}
	lower := map[string]string{}

	for _, f := range files {
		segs := strings.Split(f, "/")
		// Every path prefix: directories once, then the file.
		for i := range segs {
			prefix := strings.Join(segs[:i+1], "/")
			if seen[prefix] {
				continue
			}
			seen[prefix] = true
			c.name(prefix, segs[i])
			key := strings.ToLower(prefix)
			if other, ok := lower[key]; ok && other != prefix {
				c.errorf(prefix, "differs from %s only in case", other)
			} else {
				lower[key] = prefix
			}
		}
		c.file(f, segs)
	}
}

func (c *checker) name(p, seg string) {
	if scaffold.IsReserved(seg) {
		c.errorf(p, "%q is a reserved name on Windows", seg)
	}
	if strings.HasSuffix(seg, ".") || strings.HasSuffix(seg, " ") {
		c.errorf(p, "names must not end with a dot or a space")
	}
	for _, r := range seg {
		if r > 127 {
			c.warnf(p, "name contains non-ASCII characters")
			break
		}
	}
	if under(p, "apps") || p == "apps" || under(p, ".github") || !strings.Contains(p, "/") {
		return
	}
	stem := seg
	if i := strings.IndexByte(stem, '.'); i > 0 {
		stem = stem[:i]
	}
	if !kebab.MatchString(seg) && !upperStems[stem] && !projectFiles[seg] && !runnerFiles[seg] {
		c.warnf(p, "use lowercase kebab-case names outside apps/")
	}
}

func (c *checker) file(f string, segs []string) {
	base := segs[len(segs)-1]
	if len(f) > 120 {
		c.warnf(f, "path is longer than 120 characters")
	}
	if fi, err := os.Stat(c.p.Path(f)); err == nil && fi.Size() > maxFileSize {
		c.warnf(f, "larger than 10 MB; ask the user about Git LFS before committing")
	}
	if base == ".env" || (strings.HasPrefix(base, ".env.") && base != ".env.example") ||
		strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") ||
		strings.HasSuffix(base, ".p12") || strings.HasSuffix(base, ".pfx") {
		c.errorf(f, "looks like a secret and would be committed")
	}

	isProject := projectFiles[base] || strings.HasSuffix(base, ".csproj") ||
		strings.HasSuffix(base, ".fsproj") || strings.HasSuffix(base, ".sln")
	if isProject {
		ok := (segs[0] == "apps" && len(segs) >= 3) ||
			((under(f, "work/scripts") || under(f, "work/tests")) && len(segs) >= 4)
		if !ok {
			c.errorf(f, "project files belong in apps/<name>/ or a helper project under work/scripts/ or work/tests/")
		}
	}
	if runnerFiles[base] && segs[0] != "apps" {
		c.errorf(f, "task runner files are not allowed outside apps/; use app task tables")
	}

	switch {
	case segs[0] == "apps" && len(segs) == 2 && base != "README.md":
		c.errorf(f, "apps/ only holds app directories and README.md")
	case segs[0] == "docs" && len(segs) > 3:
		c.errorf(f, "category directories in docs/ have no subdirectories")
	case segs[0] == "work" && !inUnstored(f):
		// Levels of subdirectories below the deepest skeleton directory.
		deepest := 0
		for i := 1; i < len(segs); i++ {
			if c.skeleton[strings.Join(segs[:i], "/")] {
				deepest = i
			}
		}
		if extra := len(segs) - 1 - deepest; extra > 2 {
			c.warnf(f, "nested %d levels below %s; the limit is 2", extra, strings.Join(segs[:deepest], "/"))
		}
	}
}

func (c *checker) emptyDirs() {
	skip := map[string]bool{"node_modules": true, "target": true, ".venv": true, "__pycache__": true, ".git": true}
	for _, top := range []string{"docs", "work"} {
		_ = filepath.WalkDir(c.p.Path(top), func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(c.p.Root, p)
			rel = filepath.ToSlash(rel)
			if skip[d.Name()] || contains(local, d.Name()) || contains(unstored, rel) {
				return filepath.SkipDir
			}
			entries, err := os.ReadDir(p)
			if err == nil && len(entries) == 0 {
				c.errorf(rel, "empty directory")
			}
			return nil
		})
	}
}

func (c *checker) apps() {
	apps, err := c.p.Apps()
	if err != nil {
		return
	}
	readme, _ := os.ReadFile(c.p.Path("apps/README.md"))
	for _, name := range apps {
		rel := "apps/" + name
		if err := scaffold.ValidateAppName(name); err != nil {
			c.warnf(rel, "%v", err)
		}
		a, err := tasks.Load(name, c.p.AppDir(name), "")
		if err != nil {
			c.errorf(rel, "%v", err)
			continue
		}
		if !a.HasAgents {
			c.errorf(rel+"/AGENTS.md", "every app needs an AGENTS.md (create apps with wok new)")
			continue
		}
		if !a.HasTable {
			c.warnf(rel+"/AGENTS.md", "no task table (| task | command | windows |)")
		}
		if !bytes.Contains(readme, []byte("| "+name+" |")) && !bytes.Contains(readme, []byte("|"+name+"|")) {
			c.warnf(rel, "not listed in apps/README.md")
		}
	}
}

func (c *checker) build() {
	entries, err := os.ReadDir(c.p.Path("build"))
	if err != nil {
		return
	}
	for _, e := range entries {
		n := e.Name()
		if n == "README.md" || junk[n] {
			continue
		}
		if !e.IsDir() {
			c.warnf("build/"+n, "build/ holds one directory per app")
		} else if !c.p.HasApp(n) {
			c.warnf("build/"+n, "no app named %s", n)
		}
	}
}

func (c *checker) tmp() {
	entries, err := os.ReadDir(c.p.Path("work/tmp"))
	if err != nil {
		return
	}
	var left []string
	for _, e := range entries {
		if n := e.Name(); n != "README.md" && n != "worktrees" && !junk[n] {
			left = append(left, n)
		}
	}
	if len(left) > 0 {
		c.warnf("work/tmp", "left over from earlier tasks: %s (delete each when its task ends)", strings.Join(left, ", "))
	}
}
