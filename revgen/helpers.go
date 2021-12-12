package revgen

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func getGoRootDir() string {
	var goRootPath string
	// walk files until we find a .go file
	//  then run "go list -f '{{.Root}}'" to get the go root dir
	currDir, err := os.Getwd()
	check(err)
	err = filepath.WalkDir(currDir, func(path string, entry os.DirEntry, err error) error {
		if filepath.Ext(entry.Name()) == ".go" {
			out, err := runCmd("go list -f '{{.Root}}'", filepath.Dir(path))
			if err != nil {
				return err
			}

			// unqoute and trim new line from cmd output
			goRootPath = string(out[1 : len(out)-2])
			return errors.New("done")
		}
		return nil
	})

	// if no error : it means the walk completed and didn't find any go files
	if err == nil {
		panic("no go files found")
	}

	if err.Error() != "done" {
		panic(err)
	}

	return goRootPath
}

func getHash(rootPath string, config *GenConfig) (string, error) {
	if len(config.GenDeps) == 0 {
		return "", errors.New("missing deps")
	}

	var files []string
	for _, dep := range config.GenDeps {
		matches, err := filepath.Glob(filepath.Join(rootPath, dep))
		check(err)
		files = append(files, matches...)
	}

	if len(files) == 0 {
		return "", errors.New("empty deps")
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

func runCmd(cmdStr string, dir string) (string, error) {
	var output bytes.Buffer
	args := strings.Split(cmdStr, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
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
