package revgen

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func (a *App) getConfigMap() (*ConfigMap, error) {
	filename := filepath.Join(a.RootPath, a.ConfigFileName)
	reader, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error reading %s", filename)
	}
	return a.readConfigMap(reader)
}

func (a *App) readConfigMap(reader io.Reader) (*ConfigMap, error) {
	var config Config
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error reading %s", a.ConfigFileName)
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

func (a *App) writeConfigMap(configMap *ConfigMap) {
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
	writeYamlFile(filepath.Join(a.RootPath, a.ConfigFileName), config)
}

func (a *App) readSumMap(configMap *ConfigMap) StatusMap {
	sumList := Status{}
	sumMap := make(StatusMap)
	err := readYamlFile(filepath.Join(a.RootPath, a.SumFileName), &sumList)
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

func (a *App) writeSumMap(sumMap StatusMap) {
	sumList := make(Status, len(sumMap))
	i := 0
	for _, config := range sumMap {
		sumList[i] = config
		i++
	}
	sort.Sort(sumList)
	writeYamlFile(filepath.Join(a.RootPath, a.SumFileName), sumList)
}

func (a *App) getGoGenInfo() *ConfigMap {
	scanFileForGenerate := func(path string) []Key {
		file, err := os.Open(path)
		check(err)
		defer file.Close()

		var info []Key
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			if strings.HasPrefix(scanner.Text(), "//go:generate") {
				fields := strings.SplitN(scanner.Text(), " ", 2)

				relpath, err := filepath.Rel(a.RootPath, path)
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
	err := filepath.WalkDir(a.RootPath, func(path string, file os.DirEntry, err error) error {
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
