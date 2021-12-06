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

func readConfigMap(rootPath string) (GenMap, error) {
	var genList GenList
	err := readYamlFile(filepath.Join(rootPath, configFileName), &genList)
	if err != nil {
		return nil, fmt.Errorf("error reading %s", configFileName)
	}
	genMap := make(GenMap)
	for _, config := range genList.Configs {
		key := Key{
			FilePath: config.FilePath,
			GenCmd:   config.GenCmd,
		}
		genMap[key] = config
	}
	return genMap, nil
}

func writeConfigMap(rootPath string, genMap GenMap) {
	genList := GenList{
		Configs: make([]*GenConfig, len(genMap)),
	}
	i := 0
	for _, config := range genMap {
		genList.Configs[i] = config
		i++
	}
	sort.Sort(genList)
	writeYamlFile(filepath.Join(rootPath, configFileName), genList)
}

func readSumMap(rootPath string, genMap GenMap) SumMap {
	sumList := SumList{}
	sumMap := make(SumMap)
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
	for _, config := range genMap {
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

func writeSumMap(rootPath string, sumMap SumMap) {
	sumList := make(SumList, len(sumMap))
	i := 0
	for _, config := range sumMap {
		sumList[i] = config
		i++
	}
	sort.Sort(sumList)
	writeYamlFile(filepath.Join(rootPath, sumFileName), sumList)
}

func getGoGenInfo(rootPath string) GenMap {
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

	info := make(GenMap)
	err := filepath.WalkDir(rootPath, func(path string, file os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(file.Name()) == ".go" {
			genKeyList := scanFileForGenerate(path)
			for _, key := range genKeyList {
				info[key] = &GenConfig{
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
