package revgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	errMissingConfig = errors.New("no valid config file found, run 'revgen init' to create one")
)

// Init creates .revgen.yml config and populated it with
//  all the go:generate commands found in the codebase
func (a *App) Init(c *cli.Context) error {
	cwd, err := os.Getwd()
	check(err)
	a.rootPath = &cwd

	fmt.Printf("initializing %s\n", a.ConfigFileName)
	configMap := &ConfigMap{
		Configs: make(map[Key]*GenConfig),
	}
	a.update(configMap)
	return nil
}

// Update walks the codebase looking for go:generate lines
//  Each new go:generate is added to the config file
//  Removed go:generate lines are removed from the config file
func (a *App) Update(c *cli.Context) error {
	configMap := a.getConfigMap()
	if configMap == nil {
		return errMissingConfig
	}
	a.update(configMap)
	return nil
}

// Generate goes through all the configured go:generate commands.
//  For each command, a new hash is computed from the gen deps,
//  if this hash is different than the stored hash, the generate
//  command is run, and the new hash is stored in the sum file.
func (a *App) Generate(c *cli.Context) error {
	configMap := a.getConfigMap()
	if configMap == nil {
		return errMissingConfig
	}

	if configMap.AutoUpdate {
		a.update(configMap)
	}

	sumMap := a.readSumMap(configMap)
	if c.Bool("force") {
		for _, c := range sumMap {
			c.HashDeps = ""
		}
	}

	for key, config := range configMap.Configs {
		runGen := false
		sum := sumMap[key]
		currHash, err := getHash(*a.rootPath, "gen", config.GenDeps)
		if err != nil {
			fmt.Printf("%s:\n  %s\n  - error:%s\n", key.FilePath, key.GenCmd, err)
			runGen = true
		} else if sum.HashDeps != currHash {
			fmt.Printf("%s:\n  %s\n", key.FilePath, key.GenCmd)
			runGen = true
		} else {
			if a.debug {
				fmt.Printf("> %s: %s: match\n", key.FilePath, key.GenCmd)
			}
		}

		if runGen {
			path := filepath.Join(*a.rootPath, filepath.Dir(key.FilePath))
			err = runGenCmd(key.GenCmd, path)
			if err != nil {
				return err
			}
			sum.HashDeps = currHash
			if len(config.GenFiles) > 0 {
				filesHash, err := getHash(*a.rootPath, "file", config.GenFiles)
				check(err)
				sum.HashFiles = filesHash
			}
		}
	}

	a.writeSumMap(sumMap)
	return nil
}

// Check runs through all the genreate commands in the .revgen.yml config.
//  For each config check makes sure that the generate deps and file deps
//  match the current files in the codebase.
func (a *App) Check(c *cli.Context) error {
	configMap := a.getConfigMap()
	if configMap == nil {
		return errMissingConfig
	}

	sumMap := a.readSumMap(configMap)

	ungeneratedCode := false
	tamperedCode := false

	var messages strings.Builder
	for key, config := range configMap.Configs {
		sum := sumMap[key]
		currGenHash, err := getHash(*a.rootPath, "gen", config.GenDeps)
		if err != nil {
			messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, err))
		} else if sum.HashDeps != currGenHash {
			messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, "gen hash mismatch"))
			ungeneratedCode = true
		}

		if len(config.GenFiles) > 0 {
			currFilesHash, err := getHash(*a.rootPath, "file", config.GenFiles)
			if err != nil {
				messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, err))
			} else if sum.HashFiles != currFilesHash {
				messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, "file hash mismatch"))
				tamperedCode = true
			}
		}
	}

	errs := messages.String()
	if errs != "" {
		errorPrefix := ""
		if ungeneratedCode {
			errorPrefix += "*******************************************\n"
			errorPrefix += "* ungenerated code detected\n"
			errorPrefix += "* - did you forget to run 'revgen'?\n"
			errorPrefix += "* - did you forget to check in .revgen.yml?\n"
			errorPrefix += "*******************************************\n"
		}
		if tamperedCode {
			errorPrefix += "**************************************************\n"
			errorPrefix += "* tampered generated code detected\n"
			errorPrefix += "* - did you accidently edit some generated code?\n"
			errorPrefix += "**************************************************\n"
		}
		return errors.New(errorPrefix + errs)
	}
	return nil
}

func (a *App) update(configMap *ConfigMap) {
	gogenMap := a.getGoGenInfo()

	for key, genConfig := range gogenMap.Configs {
		if _, found := configMap.Configs[key]; !found {
			fmt.Printf("added %s:\n  %s\n", key.FilePath, key.GenCmd)
			configMap.Configs[key] = genConfig
		}
	}

	for key := range configMap.Configs {
		if _, found := gogenMap.Configs[key]; !found {
			fmt.Printf("removed %s:\n  %s\n", key.FilePath, key.GenCmd)
			delete(configMap.Configs, key)
		}
	}

	a.writeConfigMap(configMap)
}
