package revgen

import (
	"path/filepath"
	"testing"
)

func makeTestApp(path string) *App {
	app := MakeApp("test")
	app.RootPath = filepath.Join(getGoRootDir(), path)
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
