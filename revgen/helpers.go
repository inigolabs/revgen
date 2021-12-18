package revgen

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}
