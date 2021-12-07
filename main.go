package main

import (
	"fmt"
	"os"

	"github.com/inigolabs/revgen/revgen"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	cli.AppHelpTemplate = `NAME:
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

	app := &cli.App{
		Name:                 "revgen",
		Usage:                "Only run go:generate when code changes",
		HideVersion:          true,
		Version:              fmt.Sprintf("%s : %s : %s\n", version, date, commit[:8]),
		EnableBashCompletion: true,
		ExitErrHandler: func(context *cli.Context, err error) {
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
		Authors: []*cli.Author{
			{
				Name:  "Eitan Joffe",
				Email: "eitan@inigolabs.com",
			},
		},
		Action: revgen.Generate,
		Commands: []*cli.Command{
			{
				Name:   "update",
				Usage:  "Update revgen config file",
				Action: revgen.Update,
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

	app.Run(os.Args)
}
