package revgen

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

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
			c.Hash = ""
		}
	}

	for key, config := range configMap.Configs {
		sum := sumMap[key]
		currHash, err := getHash(a.RootPath, config)
		if err != nil {
			fmt.Printf("%s:\n  %s\n  - error:%s\n", key.FilePath, key.GenCmd, err)
			path := filepath.Join(a.RootPath, filepath.Dir(key.FilePath))
			_, err = runCmd(key.GenCmd, path)
			check(err)
			sum.Hash = currHash
		} else if sum.Hash != currHash {
			fmt.Printf("%s:\n  %s\n", key.FilePath, key.GenCmd)
			path := filepath.Join(a.RootPath, filepath.Dir(key.FilePath))
			_, err := runCmd(key.GenCmd, path)
			check(err)
			sum.Hash = currHash
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
		currHash, err := getHash(a.RootPath, config)
		if err != nil {
			messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, err))
		} else if sum.Hash != currHash {
			messages.WriteString(fmt.Sprintf("%s:\n  %s\n  - error: %s\n", key.FilePath, key.GenCmd, "hash mismatch"))
		}
	}

	errs := messages.String()
	if errs != "" {
		return errors.New(errs)
	}
	return nil
}
