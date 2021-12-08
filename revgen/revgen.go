package revgen

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func Update(c *cli.Context) error {
	rootPath := getGoRootDir()

	gogenMap := getGoGenInfo(rootPath)
	configMap, err := readConfigMap(rootPath)
	if err != nil {
		fmt.Printf("initializing %s\n", configFileName)
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

	writeConfigMap(rootPath, configMap)
	return nil
}

func Generate(c *cli.Context) error {
	rootPath := getGoRootDir()
	configMap, err := readConfigMap(rootPath)
	if err != nil {
		return fmt.Errorf("%s not found, run 'revgen update' to create one", configFileName)
	}

	if configMap.AutoUpdate {
		Update(c)
	}

	sumMap := readSumMap(rootPath, configMap)
	if c.Bool("force") {
		for _, c := range sumMap {
			c.Hash = ""
		}
	}

	for key, config := range configMap.Configs {
		sum := sumMap[key]
		currHash, err := getHash(rootPath, config)
		if err != nil {
			fmt.Printf("%s:\n  %s\n  - error:%s\n", key.FilePath, key.GenCmd, err)
			path := filepath.Join(rootPath, filepath.Dir(key.FilePath))
			_, err = runCmd(key.GenCmd, path)
			check(err)
			sum.Hash = currHash
		} else if sum.Hash != currHash {
			fmt.Printf("%s:\n  %s\n", key.FilePath, key.GenCmd)
			path := filepath.Join(rootPath, filepath.Dir(key.FilePath))
			_, err := runCmd(key.GenCmd, path)
			check(err)
			sum.Hash = currHash
		}
	}

	writeSumMap(rootPath, sumMap)
	return nil
}
