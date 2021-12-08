package revgen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	configFileName = ".revgen.yml"
	sumFileName    = ".revgen.sum"
)

func readConfigMap(rootPath string) (*ConfigMap, error) {
	var config Config
	err := readYamlFile(filepath.Join(rootPath, configFileName), &config)
	if err != nil {
		return nil, fmt.Errorf("error reading %s", configFileName)
	}
	genMap := &ConfigMap{
		AutoUpdate: config.AutoUpdate,
		Configs:    make(map[Key]*GenConfig),
	}
	for _, c := range config.Configs {
		key := Key{
			FilePath: c.FilePath,
			GenCmd:   c.GenCmd,
		}
		genMap.Configs[key] = c
	}
	return genMap, nil
}

func writeConfigMap(rootPath string, configMap *ConfigMap) {
	config := Config{
		AutoUpdate: configMap.AutoUpdate,
		Configs:    make([]*GenConfig, len(configMap.Configs)),
	}
	i := 0
	for _, c := range configMap.Configs {
		config.Configs[i] = c
		i++
	}
	sort.Sort(config)
	writeYamlFile(filepath.Join(rootPath, configFileName), config)
}

func readSumMap(rootPath string, configMap *ConfigMap) StatusMap {
	sumList := Status{}
	sumMap := make(StatusMap)
	err := readYamlFile(filepath.Join(rootPath, sumFileName), &sumList)
	if err == nil {
		for _, config := range sumList {
			key := Key{
				FilePath: config.FilePath,
				GenCmd:   config.GenCmd,
			}
			sumMap[key] = config
		}
	}
	for _, config := range configMap.Configs {
		key := Key{
			FilePath: config.FilePath,
			GenCmd:   config.GenCmd,
		}
		if _, found := sumMap[key]; !found {
			sumMap[key] = &SumConfig{
				FilePath: config.FilePath,
				GenCmd:   config.GenCmd,
				Hash:     "",
			}
		}
	}
	return sumMap
}

func writeSumMap(rootPath string, sumMap StatusMap) {
	sumList := make(Status, len(sumMap))
	i := 0
	for _, config := range sumMap {
		sumList[i] = config
		i++
	}
	sort.Sort(sumList)
	writeYamlFile(filepath.Join(rootPath, sumFileName), sumList)
}

func getGoGenInfo(rootPath string) *ConfigMap {
	scanFileForGenerate := func(path string) []Key {
		file, err := os.Open(path)
		check(err)
		defer file.Close()

		var info []Key
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			if strings.HasPrefix(scanner.Text(), "//go:generate") {
				fields := strings.SplitN(scanner.Text(), " ", 2)

				relpath, err := filepath.Rel(rootPath, path)
				check(err)
				info = append(info, Key{
					FilePath: relpath,
					GenCmd:   fields[1],
				})
			}
		}
		return info
	}

	info := &ConfigMap{
		Configs: make(map[Key]*GenConfig),
	}
	err := filepath.WalkDir(rootPath, func(path string, file os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(file.Name()) == ".go" {
			genKeyList := scanFileForGenerate(path)
			for _, key := range genKeyList {
				info.Configs[key] = &GenConfig{
					FilePath: key.FilePath,
					GenCmd:   key.GenCmd,
				}
			}
		}
		return nil
	})
	check(err)

	return info
}
