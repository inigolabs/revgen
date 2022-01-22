package revgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	errMissingConfig  = errors.New("no valid config file found, run 'revgen init' to create one")
	errExistingConfig = errors.New("config file found, run 'revgen update' to update it with new go::generators")
)

// Init creates .revgen.yml config and populated it with
//  all the go:generate commands found in the codebase
func (a *App) Init(c *cli.Context) error {
	cwd, err := os.Getwd()
	check(err)
	a.rootPath = &cwd

	config := a.getConfig()
	if config != nil {
		return errExistingConfig
	}

	config = &Config{
		AutoUpdate: true,
		Generators: make(map[Name]*GenConfig),
	}
	fmt.Printf("Initializing %s\n", a.ConfigFileName)
	a.update(config)
	return nil
}

// Update walks the codebase looking for go:generate lines
//  Each new go:generate is added to the config file
//  Removed go:generate lines are removed from the config file
func (a *App) Update(c *cli.Context) error {
	config := a.getConfig()
	if config == nil {
		return errMissingConfig
	}
	a.update(config)
	return nil
}

// Generate goes through all the configured go:generate commands.
//  For each command, a new hash is computed from the gen deps,
//  if this hash is different than the stored hash, the generate
//  command is run, and the new hash is stored in the sum file.
func (a *App) Generate(c *cli.Context) error {
	config := a.getConfig()
	if config == nil {
		return errMissingConfig
	}

	if config.AutoUpdate {
		a.update(config)
	}

	status := a.getStatus()

	if c.Bool("force") {
		for _, s := range status {
			s.InputsHash = ""
		}
	}

	for name, config := range config.Generators {
		runGen := false
		sum := status[name]
		currHash, err := getHash(*a.rootPath, "gen", config.Inputs)
		if err != nil {
			fmt.Printf("error %s:%s\n", name, err)
			runGen = true
		} else if sum == nil || sum.InputsHash != currHash {
			fmt.Printf("%s: %s:\n  %s\n", name, config.FilePath, config.GenCmd)
			runGen = true
		} else {
			if a.debug {
				fmt.Printf("> %s: hash match\n", name)
			}
		}

		if runGen {
			path := filepath.Join(*a.rootPath, filepath.Dir(config.FilePath))
			err = runGenCmd(config.GenCmd, path)
			if err != nil {
				return err
			}

			// update status hashes
			var outputsHash string
			if len(config.Outputs) > 0 {
				outputsHash, err = getHash(*a.rootPath, "file", config.Outputs)
				check(err)
			}
			status[name] = &SumConfig{
				InputsHash:  currHash,
				OutputsHash: outputsHash,
			}
		}
	}

	writeYamlFile(filepath.Join(*a.rootPath, a.SumFileName), status)
	return nil
}

// Check runs through all the genreate commands in the .revgen.yml config.
//  For each config check makes sure that the generate deps and file deps
//  match the current files in the codebase.
func (a *App) Check(c *cli.Context) error {
	config := a.getConfig()
	if config == nil {
		return errMissingConfig
	}

	status := a.getStatus()

	ungeneratedCode := false
	tamperedCode := false

	var messages strings.Builder
	for name, config := range config.Generators {
		sum := status[name]
		currGenHash, err := getHash(*a.rootPath, "gen", config.Inputs)
		if err != nil {
			messages.WriteString(fmt.Sprintf("error: %s: %s\n", name, err))
		} else if sum.InputsHash != currGenHash {
			messages.WriteString(fmt.Sprintf("error: %s: %s\n", name, "gen hash mismatch"))
			ungeneratedCode = true
		}

		if len(config.Outputs) > 0 {
			currFilesHash, err := getHash(*a.rootPath, "file", config.Outputs)
			if err != nil {
				messages.WriteString(fmt.Sprintf("error: %s: %s\n", name, err))
			} else if sum.OutputsHash != currFilesHash {
				messages.WriteString(fmt.Sprintf("error: %s: %s\n", name, "file hash mismatch"))
				tamperedCode = true
			}
		}
	}

	errs := messages.String()
	if errs != "" {
		errorPrefix := ""
		if ungeneratedCode {
			errorPrefix += "*******************************************\n"
			errorPrefix += "* ungenerated code detected\n"
			errorPrefix += "* - did you forget to run 'revgen'?\n"
			errorPrefix += "* - did you forget to check in .revgen.yml?\n"
			errorPrefix += "*******************************************\n"
		}
		if tamperedCode {
			errorPrefix += "**************************************************\n"
			errorPrefix += "* tampered generated code detected\n"
			errorPrefix += "* - did you accidently edit some generated code?\n"
			errorPrefix += "**************************************************\n"
		}
		return errors.New(errorPrefix + errs)
	}
	return nil
}

func (a *App) update(config *Config) {

	configMap := make(map[GenInfo]Name)
	for name, genConfig := range config.Generators {
		info := GenInfo{
			FilePath: genConfig.FilePath,
			GenCmd:   genConfig.GenCmd,
		}
		configMap[info] = name
	}

	infoMap := make(map[GenInfo]struct{})

	infoList := a.getGoGenInfo()
	for _, info := range infoList {
		infoMap[*info] = struct{}{}
		if _, found := configMap[*info]; !found {
			// new generator
			fmt.Printf("Enter name for generator:\n")
			fmt.Printf(" file: %s\n", info.FilePath)
			fmt.Printf("  cmd: %s\n", info.GenCmd)
			fmt.Print(" name: ")
			var name string
			if _, err := fmt.Scan(&name); err != nil {
				check(err)
			}
			fmt.Println()
			config.Generators[Name(name)] = &GenConfig{
				FilePath: info.FilePath,
				GenCmd:   info.GenCmd,
			}
		}
	}

	for key, name := range configMap {
		if _, found := infoMap[key]; !found {
			fmt.Printf("removed %s\n", name)
			delete(config.Generators, name)
		}
	}

	writeYamlFile(filepath.Join(*a.rootPath, a.ConfigFileName), config)
}
