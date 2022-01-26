package revgen

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func getHash(rootPath string, hashType string, globs []string) (string, error) {
	if len(globs) == 0 {
		return "", fmt.Errorf("missing %s deps", hashType)
	}

	var matchedFiles []string
	var files []string

	for _, dep := range globs {
		matches, err := filepath.Glob(filepath.Join(rootPath, dep))
		check(err)
		matchedFiles = append(matchedFiles, matches...)
	}

	if len(matchedFiles) == 0 {
		return "", fmt.Errorf("empty %s deps", hashType)
	}

	for _, file := range matchedFiles {
		fileInfo, err := os.Stat(file)
		check(err)

		if fileInfo.IsDir() {
			err := filepath.Walk(file,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					fileInfo, err := os.Stat(path)
					if err != nil {
						return err
					}
					if !fileInfo.IsDir() {
						files = append(files, path)
					}
					return nil
				})
			check(err)
		} else {
			files = append(files, file)
		}
	}

	hash := md5.New()
	for _, filename := range files {
		reader, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
		check(err)
		_, err = io.Copy(hash, reader)
		check(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// orderConfigs returns a list of config names making sure that
//  any generator that depends on another generator comes after
//  it in the list.
func orderConfigs(config *Config) []Name {
	var orderedList []Name
	visited := make(map[Name]struct{})

	keys := make([]string, len(config.Generators))
	index := 0
	for k := range config.Generators {
		keys[index] = string(k)
		index++
	}
	sort.Strings(keys)

	for _, key := range keys {
		name := Name(key)
		if _, found := visited[name]; found {
			continue
		}

		gen := config.Generators[name]
		if gen.GenDep == "" {
			// if not dependancy we just -> add name to the end
			orderedList = append(orderedList, name)
			visited[name] = struct{}{}
		} else {
			// else -> add all non visited dependencies
			localList := []Name{name}
			depName := Name(gen.GenDep)
			for {
				if depName == "" {
					break
				}
				if _, found := visited[depName]; found {
					break
				}

				localList = append([]Name{depName}, localList...)
				dep := config.Generators[depName]
				if dep.GenDep == "" {
					break
				}
				depName = dep.GenDep
			}

			for _, localName := range localList {
				visited[localName] = struct{}{}
			}
			orderedList = append(orderedList, localList...)
		}
	}

	return orderedList
}

func runCmd(cmdStr string, dir *string) (string, error) {
	var output bytes.Buffer
	args := strings.Split(cmdStr, " ")
	cmd := exec.Command(args[0], args[1:]...)
	if dir != nil {
		cmd.Dir = *dir
	}
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Env = os.Environ()
	err := cmd.Run()
	return output.String(), err
}

func runGenCmd(cmdStr string, dir string) error {
	args := strings.Split(cmdStr, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	err := cmd.Run()
	return err
}

func readYamlFile(filename string, dataPtr interface{}) error {
	reader, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(dataPtr)
	if err != nil {
		return err
	}
	return nil
}

func writeYamlFile(filename string, dataPtr interface{}) {
	writer, err := os.Create(filename)
	check(err)
	encoder := yaml.NewEncoder(writer)
	err = encoder.Encode(dataPtr)
	check(err)
}

func pretty(obj interface{}) {
	objectString, err := json.MarshalIndent(obj, "", "  ")
	check(err)
	fmt.Println(string(objectString))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
