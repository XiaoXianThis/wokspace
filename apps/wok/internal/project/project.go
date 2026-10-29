// Package project locates a wokspace and its apps.
package project

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LayoutVersion is the version of the layout spec this wok implements. It
// matches the <!-- wokspace-layout: N --> marker in the template's AGENTS.md.
const LayoutVersion = 1

// ErrNotFound means no wokspace root was found.
var ErrNotFound = errors.New("not inside a wokspace")

// Project is a wokspace on disk.
type Project struct {
	Root string
}

// Find walks up from start to the nearest directory that has both an
// AGENTS.md file and an apps/ directory.
func Find(start string) (*Project, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	for {
		if isFile(filepath.Join(dir, "AGENTS.md")) && isDir(filepath.Join(dir, "apps")) {
			return &Project{Root: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotFound
		}
		dir = parent
	}
}

func (p *Project) Path(slashPath string) string {
	return filepath.Join(p.Root, filepath.FromSlash(slashPath))
}

func (p *Project) AppsDir() string           { return filepath.Join(p.Root, "apps") }
func (p *Project) AppDir(name string) string { return filepath.Join(p.Root, "apps", name) }
func (p *Project) OutDir(name string) string { return filepath.Join(p.Root, "build", name) }
func (p *Project) HasApp(name string) bool   { return isDir(p.AppDir(name)) }

// Apps lists app directory names in apps/, sorted.
func (p *Project) Apps() ([]string, error) {
	entries, err := os.ReadDir(p.AppsDir())
	if err != nil {
		return nil, err
	}
	var apps []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			apps = append(apps, e.Name())
		}
	}
	sort.Strings(apps)
	return apps, nil
}

// AppAt returns the app that dir is inside, or "".
func (p *Project) AppAt(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	rel, err := filepath.Rel(p.AppsDir(), abs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ""
	}
	name := strings.Split(filepath.ToSlash(rel), "/")[0]
	if p.HasApp(name) {
		return name
	}
	return ""
}

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
