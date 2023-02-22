package revgen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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

func (a *App) getStatus() (Status, error) {
	filename := filepath.Join(*a.rootPath, a.SumFileName)
	var status Status
	ymlerr := readYamlFile(filename, &status)
	if ymlerr != nil {
		// try to remove conflicts
		fmt.Println("trying to autoremove conflicts")
		reader, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return nil, ymlerr
		}
		scanner := bufio.NewScanner(reader)
		var builder strings.Builder
		inConflict := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "<<<<<<< HEAD") {
				inConflict = true
			}

			if !inConflict {
				builder.WriteString(line + "\n")
			}

			if strings.HasPrefix(line, ">>>>>>>") {
				inConflict = false
			}
		}

		decoder := yaml.NewDecoder(strings.NewReader(builder.String()))
		err = decoder.Decode(&status)
		if err != nil {
			return nil, ymlerr
		}
	}
	return status, nil
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
