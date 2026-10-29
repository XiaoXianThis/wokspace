// Package tasks resolves the commands an app's tasks run: from the task table
// in the app's AGENTS.md first, then from its project files.
package tasks

import (
	"strings"
)

// Entry is one row of a task table.
type Entry struct {
	Command string
	Windows string
}

// Table is the first task table in a Markdown document.
type Table struct {
	Found   bool
	Order   []string
	Entries map[string]Entry
}

// ParseTable finds the first table whose header starts with "task | command"
// outside fenced code blocks. An optional "windows" column overrides the
// command on Windows. "\|" stands for a literal pipe.
func ParseTable(md string) Table {
	t := Table{Entries: map[string]Entry{}}
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	fence := ""
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if fence != "" {
			if closesFence(line, fence) {
				fence = ""
			}
			continue
		}
		if f := fenceMarker(line); f != "" {
			fence = f
			continue
		}
		if !strings.HasPrefix(line, "|") {
			continue
		}
		header := splitRow(line)
		if len(header) < 2 || !strings.EqualFold(header[0], "task") || !strings.EqualFold(header[1], "command") {
			continue
		}
		if i+1 >= len(lines) || !isSeparator(strings.TrimSpace(lines[i+1])) {
			continue
		}
		winCol := -1
		for c, h := range header {
			if strings.EqualFold(h, "windows") {
				winCol = c
			}
		}
		t.Found = true
		for j := i + 2; j < len(lines); j++ {
			row := strings.TrimSpace(lines[j])
			if !strings.HasPrefix(row, "|") {
				break
			}
			cells := splitRow(row)
			name := strings.ToLower(unquote(cells[0]))
			if name == "" {
				continue
			}
			var e Entry
			if len(cells) > 1 {
				e.Command = unquote(cells[1])
			}
			if winCol >= 0 && winCol < len(cells) {
				e.Windows = unquote(cells[winCol])
			}
			if _, dup := t.Entries[name]; !dup {
				t.Order = append(t.Order, name)
			}
			t.Entries[name] = e
		}
		return t
	}
	return t
}

func fenceMarker(line string) string {
	for _, ch := range []string{"`", "~"} {
		n := 0
		for n < len(line) && line[n:n+1] == ch {
			n++
		}
		if n >= 3 {
			return line[:n]
		}
	}
	return ""
}

func closesFence(line, fence string) bool {
	f := fenceMarker(line)
	return f != "" && f[0] == fence[0] && len(f) >= len(fence) && strings.TrimSpace(line[len(f):]) == ""
}

func splitRow(line string) []string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "|")
	if strings.HasSuffix(s, "|") && !strings.HasSuffix(s, `\|`) {
		s = s[:len(s)-1]
	}
	var cells []string
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) && s[i+1] == '|' {
			b.WriteByte('|')
			i++
			continue
		}
		if c == '|' {
			cells = append(cells, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		}
		b.WriteByte(c)
	}
	return append(cells, strings.TrimSpace(b.String()))
}

func isSeparator(line string) bool {
	if !strings.HasPrefix(line, "|") {
		return false
	}
	for _, c := range splitRow(line) {
		c = strings.Trim(c, ":")
		if c == "" || strings.Trim(c, "-") != "" {
			return false
		}
	}
	return true
}

// unquote removes inline-code backticks around a cell.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	for len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}
