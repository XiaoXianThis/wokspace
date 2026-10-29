// Package ui formats terminal output.
package ui

import (
	"os"
)

// Enabled reports whether ANSI colors are written.
var Enabled = detect()

func detect() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	for _, f := range []*os.File{os.Stdout, os.Stderr} {
		fi, err := f.Stat()
		if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
			return false
		}
	}
	return enableVT(os.Stdout) && enableVT(os.Stderr)
}

func paint(code, s string) string {
	if !Enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func Bold(s string) string   { return paint("1", s) }
func Dim(s string) string    { return paint("2", s) }
func Red(s string) string    { return paint("31", s) }
func Green(s string) string  { return paint("32", s) }
func Yellow(s string) string { return paint("33", s) }
func Cyan(s string) string   { return paint("36", s) }

var palette = []string{"36", "35", "33", "32", "34", "96", "95", "93"}

// Label colors s with the i-th color of a fixed palette, for telling
// parallel processes apart.
func Label(i int, s string) string { return paint(palette[i%len(palette)], s) }
