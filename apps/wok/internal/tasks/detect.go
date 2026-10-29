package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// detected is the set of tasks inferred from one project file.
type detected struct {
	source string
	order  []string
	cmds   map[string]string
}

func newDetected(source string) *detected {
	return &detected{source: source, cmds: map[string]string{}}
}

func (d *detected) add(task, cmd string) {
	if _, ok := d.cmds[task]; ok {
		return
	}
	d.order = append(d.order, task)
	d.cmds[task] = cmd
}

// detectAll runs every detector in priority order. The first detector that
// defines a task wins.
func detectAll(dir, goos string) []*detected {
	var out []*detected
	for _, fn := range []func(string, string) *detected{
		detectNode, detectDeno, detectCargo, detectGo, detectDotnet, detectGradle, detectMaven, detectPython,
	} {
		if d := fn(dir, goos); d != nil {
			out = append(out, d)
		}
	}
	return out
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func detectNode(dir, _ string) *detected {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts        map[string]string `json:"scripts"`
		PackageManager string            `json:"packageManager"`
	}
	_ = json.Unmarshal(data, &pkg)
	pm := nodePackageManager(dir, pkg.PackageManager)
	d := newDetected("package.json")
	d.add("install", pm+" install")
	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		d.add(name, pm+" run "+name)
	}
	return d
}

func nodePackageManager(dir, field string) string {
	for _, pm := range []string{"pnpm", "yarn", "bun", "npm"} {
		if strings.HasPrefix(field, pm+"@") {
			return pm
		}
	}
	switch {
	case exists(dir, "pnpm-lock.yaml"):
		return "pnpm"
	case exists(dir, "yarn.lock"):
		return "yarn"
	case exists(dir, "bun.lock"), exists(dir, "bun.lockb"):
		return "bun"
	}
	return "npm"
}

func detectDeno(dir, _ string) *detected {
	data, err := os.ReadFile(filepath.Join(dir, "deno.json"))
	if err != nil {
		return nil
	}
	var cfg struct {
		Tasks map[string]json.RawMessage `json:"tasks"`
	}
	_ = json.Unmarshal(data, &cfg)
	d := newDetected("deno.json")
	d.add("install", "deno install")
	names := make([]string, 0, len(cfg.Tasks))
	for name := range cfg.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		d.add(name, "deno task "+name)
	}
	return d
}

func detectCargo(dir, _ string) *detected {
	if !exists(dir, "Cargo.toml") {
		return nil
	}
	d := newDetected("Cargo.toml")
	d.add("install", "cargo fetch")
	d.add("dev", "cargo run")
	d.add("test", "cargo test")
	d.add("build", "cargo build --release")
	return d
}

func detectGo(dir, _ string) *detected {
	if !exists(dir, "go.mod") {
		return nil
	}
	d := newDetected("go.mod")
	d.add("install", "go mod download")
	d.add("dev", "go run .")
	d.add("test", "go test ./...")
	d.add("build", `go build -o "$WOK_OUT" .`)
	return d
}

func detectDotnet(dir, _ string) *detected {
	var found string
	for _, pattern := range []string{"*.sln", "*.csproj", "*.fsproj"} {
		if m, _ := filepath.Glob(filepath.Join(dir, pattern)); len(m) > 0 {
			found = filepath.Base(m[0])
			break
		}
	}
	if found == "" {
		return nil
	}
	d := newDetected(found)
	d.add("install", "dotnet restore")
	d.add("dev", "dotnet run")
	d.add("test", "dotnet test")
	d.add("build", `dotnet publish -c Release -o "$WOK_OUT"`)
	return d
}

func detectGradle(dir, goos string) *detected {
	var g, source string
	switch {
	case exists(dir, "gradlew"), exists(dir, "gradlew.bat"):
		g, source = "./gradlew", "gradlew"
		if goos == "windows" {
			g = "gradlew.bat"
		}
	case exists(dir, "build.gradle.kts"):
		g, source = "gradle", "build.gradle.kts"
	case exists(dir, "build.gradle"):
		g, source = "gradle", "build.gradle"
	default:
		return nil
	}
	d := newDetected(source)
	d.add("test", g+" test")
	d.add("build", g+" build")
	return d
}

func detectMaven(dir, goos string) *detected {
	if !exists(dir, "pom.xml") {
		return nil
	}
	m := "mvn"
	if exists(dir, "mvnw") || exists(dir, "mvnw.cmd") {
		m = "./mvnw"
		if goos == "windows" {
			m = "mvnw.cmd"
		}
	}
	d := newDetected("pom.xml")
	d.add("install", m+" dependency:resolve")
	d.add("test", m+" test")
	d.add("build", m+" package")
	return d
}

func detectPython(dir, _ string) *detected {
	if !exists(dir, "pyproject.toml") {
		return nil
	}
	switch {
	case exists(dir, "uv.lock"):
		d := newDetected("pyproject.toml + uv.lock")
		d.add("install", "uv sync")
		d.add("test", "uv run pytest")
		d.add("build", `uv build --out-dir "$WOK_OUT"`)
		return d
	case exists(dir, "poetry.lock"):
		d := newDetected("pyproject.toml + poetry.lock")
		d.add("install", "poetry install")
		d.add("test", "poetry run pytest")
		d.add("build", "poetry build")
		return d
	}
	return nil
}
