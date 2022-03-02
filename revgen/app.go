package revgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type App struct {
	ConfigFileName string
	SumFileName    string

	rootPath *string
	cliapp   *cli.App
	debug    bool
}

func MakeApp(version string) *App {
	cli.AppHelpTemplate = template

	app := &App{
		ConfigFileName: ".revgen.yml",
		SumFileName:    ".revgen.sum",
	}

	app.rootPath = findRootDir(app.ConfigFileName)

	app.cliapp = &cli.App{
		Name:                 "revgen",
		Usage:                "Only run go:generate when code changes",
		HideVersion:          true,
		Version:              version,
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Eitan Joffe",
				Email: "eitan@inigolabs.com",
			},
		},
		Before: func(c *cli.Context) error {
			if c.IsSet("debug") {
				fmt.Printf("> root_path:%v\n", *app.rootPath)
				app.debug = true
			}
			return nil
		},
		Action: app.Generate,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force run on all generators regardless of code changes",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Print debug logs",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Init revgen config file",
				Action: app.Init,
			},
			{
				Name:   "update",
				Usage:  "Update revgen config file",
				Action: app.Update,
			},
			{
				Name:   "check",
				Usage:  "Check that all hashes match with gen file deps",
				Action: app.Check,
			},
			{
				Name:  "version",
				Usage: "Show version info",
				Action: func(c *cli.Context) error {
					return cli.Exit(c.App.Version, 0)
				},
			},
		},
	}

	return app
}

func (a *App) Run(args []string) {
	err := a.cliapp.Run(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var template string = `NAME:
{{.Name}} - {{.Usage}}

USAGE:
{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

GLOBAL OPTIONS:
{{range .VisibleFlags}}{{"\t"}}{{.}}
{{end}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ","}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}
AUTHOR: {{range .Authors}}{{ . }}{{end}}

VERSION: {{.Version}}
`

func findRootDir(configFileName string) *string {
	currDir, err := os.Getwd()
	check(err)
	_, err = runCmd("git rev-parse --show-toplevel", &currDir)
	check(err)
	filename := filepath.Join(currDir, configFileName)
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		currDir = filepath.Dir(currDir)
	}
	return &currDir
}
