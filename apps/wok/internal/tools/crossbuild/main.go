// Command crossbuild builds wok for every release platform into one
// directory, using the release artifact names (wok_<os>_<arch>[.exe]).
// It is the local equivalent of the release build, without goreleaser.
//
// Usage: go run ./internal/tools/crossbuild <out-dir> [version]
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var targets = [][2]string{
	{"darwin", "amd64"}, {"darwin", "arm64"},
	{"linux", "amd64"}, {"linux", "arm64"},
	{"windows", "amd64"}, {"windows", "arm64"},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: crossbuild <out-dir> [version]")
		os.Exit(2)
	}
	out := os.Args[1]
	version := "dev"
	if len(os.Args) > 2 {
		version = os.Args[2]
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, t := range targets {
		name := fmt.Sprintf("wok_%s_%s", t[0], t[1])
		if t[0] == "windows" {
			name += ".exe"
		}
		cmd := exec.Command("go", "build", "-trimpath",
			"-ldflags", "-s -w -X main.version="+version,
			"-o", filepath.Join(out, name), ".")
		cmd.Env = append(os.Environ(), "GOOS="+t[0], "GOARCH="+t[1], "CGO_ENABLED=0")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		fmt.Println("build", name)
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
	}
}
