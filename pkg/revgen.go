package revgen

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

const (
	configFileName = ".revgen.yml"
	sumFileName    = ".revgen.sum"
)

type GenFile struct {
	Configs []*GenConfig `json:"Configs"`
}

type GenConfig struct {
	Key     Key      `json:"Key"`
	GenDeps []string `json:"GenDeps"`
}

type SumFile struct {
	Configs []*SumConfig `json:"Configs"`
}

type SumConfig struct {
	Key  Key    `json:"Key"`
	Hash string `json:"Hash"`
}

type Key struct {
	FilePath string `json:"FilePath"`
	GenCmd   string `json:"GenCmd"`
}

type gen struct {
	rootPath    string
	genDirPath  string
	genFilePath string
	genCmd      string

	configFilePath string
	sumFilePath    string
}

func Generate(args []string) {
	rootPath := pathRepoRoot()
	genPath := pathGenFile(rootPath)
	pathGenFile(rootPath)
	g := &gen{
		rootPath:       rootPath,
		genDirPath:     genPath,
		genFilePath:    filepath.Join(genPath, os.Getenv("GOFILE")),
		genCmd:         strings.Join(args[1:], " "),
		configFilePath: filepath.Join(rootPath, configFileName),
		sumFilePath:    filepath.Join(rootPath, sumFileName),
	}

	configFile, config := g.getConfig()
	currHash := g.getCurrentHash(config)
	sumFile, sum := g.getSum()

	if currHash == "" {
		fmt.Printf("%s:%s : missing dependencies in .revgen.yml\n", g.genFilePath, g.genCmd)
		err := runGen(g.genCmd, filepath.Join(g.rootPath, g.genDirPath))
		check(err)
		sum.Hash = currHash
	} else if sum.Hash != currHash {
		fmt.Printf("%s:%s\n", g.genFilePath, g.genCmd)
		err := runGen(g.genCmd, filepath.Join(g.rootPath, g.genDirPath))
		check(err)
		sum.Hash = currHash
	}

	writeYamlFile(g.configFilePath, configFile)
	writeYamlFile(g.sumFilePath, sumFile)
}

func (g *gen) getConfig() (*GenFile, *GenConfig) {
	var genFile GenFile
	reader, err := fileReader(g.configFilePath)
	if err == nil {
		decoder := yaml.NewDecoder(reader)
		err = decoder.Decode(&genFile)
		if err != nil {
			panic(fmt.Errorf("malformed config file : %w", err))
		}
	}

	var genConfig *GenConfig
	for _, c := range genFile.Configs {
		if c.Key.GenCmd == g.genCmd && c.Key.FilePath == g.genFilePath {
			genConfig = c
		}
	}
	if genConfig == nil {
		genConfig = &GenConfig{
			Key: Key{
				FilePath: g.genFilePath,
				GenCmd:   g.genCmd,
			},
		}
		genFile.Configs = append(genFile.Configs, genConfig)
	}
	return &genFile, genConfig
}

func (g *gen) getSum() (*SumFile, *SumConfig) {
	var sumFile SumFile
	reader, err := fileReader(g.sumFilePath)
	if err == nil {
		decoder := yaml.NewDecoder(reader)
		err = decoder.Decode(&sumFile)
		if err != nil {
			panic(fmt.Errorf("malformed sum file : %w", err))
		}
	}

	var sumConfig *SumConfig
	for _, s := range sumFile.Configs {
		if s.Key.GenCmd == g.genCmd && s.Key.FilePath == g.genFilePath {
			sumConfig = s
		}
	}
	if sumConfig == nil {
		sumConfig = &SumConfig{
			Key: Key{
				FilePath: g.genFilePath,
				GenCmd:   g.genCmd,
			},
		}
		sumFile.Configs = append(sumFile.Configs, sumConfig)
	}
	return &sumFile, sumConfig
}

func (g *gen) getCurrentHash(config *GenConfig) string {
	if len(config.GenDeps) == 0 {
		return ""
	}

	var files []string
	for _, deps := range config.GenDeps {
		matches, err := filepath.Glob(deps)
		check(err)
		files = append(files, matches...)
	}

	hash := md5.New()
	for _, filename := range files {
		reader, err := fileReader(filepath.Join(g.rootPath, filename))
		check(err)
		_, err = io.Copy(hash, reader)
		check(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func pathRepoRoot() string {
	// TODO - get go root dir without using git and remove + "/go" hardcoding
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	check(err)
	return strings.TrimSpace(string(out)) + "/go"
}

func pathGenFile(rootPath string) string {
	fullPath, err := os.Getwd()
	check(err)
	relPath, err := filepath.Rel(rootPath, fullPath)
	check(err)
	return relPath
}

func runGen(cmdStr string, dir string) error {
	args := strings.Split(cmdStr, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	err := cmd.Run()
	return err
}

func fileReader(filename string) (io.Reader, error) {
	return os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
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
