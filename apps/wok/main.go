// Command wok creates and runs wokspaces: repositories whose top level is a
// workspace for agents and humans. See https://wokspace.wxian.cc.
package main

import (
	"embed"
	"io/fs"
	"os"
	"runtime/debug"
	"strings"

	"github.com/XiaoXianThis/wokspace/apps/wok/internal/cli"
)

// The skeleton that `wok init` writes. Files named dot-* become .* on disk.
//
//go:embed all:template
var templateFS embed.FS

// Set at release time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	// `go install ...@vX.Y.Z` does not apply ldflags; take the module version.
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && strings.HasPrefix(info.Main.Version, "v") {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	tmpl, err := fs.Sub(templateFS, "template")
	if err != nil {
		panic(err)
	}
	os.Exit(cli.Main(os.Args[1:], cli.Env{
		Version:  version,
		Template: tmpl,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
	}))
}
