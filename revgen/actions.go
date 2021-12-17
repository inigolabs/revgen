package revgen

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// Update walks the codebase looking for go:generate lines
//  Each new go:generate is added to the config file
//  Removed go:generate lines are removed from the config file
func (a *App) Update(c *cli.Context) error {
	gogenMap := a.getGoGenInfo()
	configMap, err := a.getConfigMap()
	if err != nil {
		fmt.Printf("initializing %s\n", a.ConfigFileName)
		configMap = &ConfigMap{
			Configs: make(map[Key]*GenConfig),
		}
	}

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
	return nil
}

// Generate goes through all the configured go:generate commands.
//  For each command, a new hash is computed from the gen deps,
//  if this hash is different than the stored hash, the generate
//  command is run, and the new hash is stored in the sum file.
func (a *App) Generate(c *cli.Context) error {
	configMap, err := a.getConfigMap()
	if err != nil {
		return fmt.Errorf("%s not found, run 'revgen update' to create one", a.ConfigFileName)
	}

	if configMap.AutoUpdate {
		a.Update(c)
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
		currHash, err := getHash(a.RootPath, "gen", config.GenDeps)
		if err != nil {
			fmt.Printf("%s:\n  %s\n  - error:%s\n", key.FilePath, key.GenCmd, err)
			runGen = true
		} else if sum.HashDeps != currHash {
			fmt.Printf("%s:\n  %s\n", key.FilePath, key.GenCmd)
			runGen = true
		}

		if runGen {
			path := filepath.Join(a.RootPath, filepath.Dir(key.FilePath))
			err = runGenCmd(key.GenCmd, path)
			if err != nil {
				return err
			}
			sum.HashDeps = currHash
			if len(config.GenFiles) > 0 {
				filesHash, err := getHash(a.RootPath, "file", config.GenFiles)
				check(err)
				sum.HashFiles = filesHash
			}
		}
	}

	a.writeSumMap(sumMap)
	return nil
}

func (a *App) Check(c *cli.Context) error {
	configMap, err := a.getConfigMap()
	if err != nil {
		return fmt.Errorf("%s not found, run 'revgen update' to create one", a.ConfigFileName)
	}

	sumMap := a.readSumMap(configMap)

	var messages strings.Builder
	for key, config := range configMap.Configs {
		sum := sumMap[key]
		currGenHash, err := getHash(a.RootPath, "gen", config.GenDeps)
		if err != nil {
			messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, err))
		} else if sum.HashDeps != currGenHash {
			messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, "gen hash mismatch"))
		}

		if len(config.GenFiles) > 0 {
			currFilesHash, err := getHash(a.RootPath, "file", config.GenFiles)
			if err != nil {
				messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, err))
			} else if sum.HashFiles != currFilesHash {
				messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, "file hash mismatch"))
			}
		}
	}

	errs := messages.String()
	if errs != "" {
		return errors.New(errs)
	}
	return nil
}
