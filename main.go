package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/inigolabs/revgen/revgen"
	"github.com/urfave/cli/v2"
)

var (
	version = ""
	commit  = ""
	date    = ""
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
		Version:              buildVersion(version, commit, date),
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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force run on all generators regardless of code changes",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Init revgen config file",
				Action: revgen.Update,
			},
			{
				Name:   "update",
				Usage:  "Update revgen config file",
				Action: revgen.Update,
			},
			{
				Name:   "check",
				Usage:  "Check that all hashes match with gen file deps",
				Action: revgen.Check,
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

func buildVersion(version, commit, date string) string {
	var result strings.Builder
	if version != "" {
		result.WriteString(fmt.Sprintf("version: %s\n", version))
	}
	if commit != "" {
		result.WriteString(fmt.Sprintf("commit: %s\n", commit))
	}
	if date != "" {
		result.WriteString(fmt.Sprintf("built at: %s", date))
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result.WriteString(fmt.Sprintf("module version: %s\n", info.Main.Version))
		result.WriteString(fmt.Sprintf("checksum: %s\n", info.Main.Sum))
	}
	result.WriteString("\n")
	result.WriteString("https://github.com/inigolabs/revgen\n")
	return result.String()
}
