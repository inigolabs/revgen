package revgen

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) getConfig() *Config {
	if a.rootPath == nil {
		return nil
	}
	filename := filepath.Join(*a.rootPath, a.ConfigFileName)
	var config Config
	err := readYamlFile(filename, &config)
	if err != nil {
		return nil
	}
	return &config
}

func (a *App) getStatus() Status {
	filename := filepath.Join(*a.rootPath, a.SumFileName)
	var status Status
	err := readYamlFile(filename, &status)
	if err != nil {
		return Status{}
	}
	return status
}

func (a *App) getGoGenInfo() []*GenInfo {
	scanFileForGenerate := func(path string) []*GenInfo {
		file, err := os.Open(path)
		check(err)
		defer file.Close()

		var infoList []*GenInfo
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			if strings.HasPrefix(scanner.Text(), "//go:generate") {
				fields := strings.SplitN(scanner.Text(), " ", 2)

				relpath, err := filepath.Rel(*a.rootPath, path)
				check(err)
				infoList = append(infoList, &GenInfo{
					FilePath: relpath,
					GenCmd:   fields[1],
				})
			}
		}
		return infoList
	}

	var infoList []*GenInfo
	err := filepath.WalkDir(*a.rootPath, func(path string, file os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(file.Name()) == ".go" {
			infoList = append(infoList, scanFileForGenerate(path)...)
		}
		return nil
	})
	check(err)

	return infoList
}
