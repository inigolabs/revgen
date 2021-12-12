package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/inigolabs/revgen/revgen"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	versionStr := buildVersion(version, commit, date)
	app := revgen.MakeApp(versionStr)
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
