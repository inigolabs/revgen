package revgen

import (
	"path/filepath"
	"strings"
	"testing"
)

func makeTestApp(path string) *App {
	app := MakeApp("test")

	out, err := runCmd("git rev-parse --show-toplevel", nil)
	check(err)
	rootDir := strings.TrimSpace(out)
	path = filepath.Join(rootDir, path)
	app.rootPath = &path
	app.ConfigFileName = "cfg.yml"
	app.SumFileName = "sum.yml"
	return app
}

func TestRevgenUpdate(t *testing.T) {
	app := makeTestApp("revgen/test")
	app.cliapp.Run([]string{"revgen", "update"})
}

func TestRevgenGenerate(t *testing.T) {
	app := makeTestApp("revgen/test")
	app.cliapp.Run([]string{"revgen"})
}

func TestRevgenCheck(t *testing.T) {
	app := makeTestApp("revgen/test")
	app.cliapp.Run([]string{"revgen", "check"})
}
