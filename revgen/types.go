package revgen

type Name string

type Config struct {
	AutoUpdate bool                `yaml:"auto_update"`
	Generators map[Name]*GenConfig `yaml:"generators"`
}

type Status map[Name]*SumConfig

type GenConfig struct {
	FilePath string   `yaml:"path"`
	GenCmd   string   `yaml:"gen_cmd"`
	GenDep   Name     `yaml:"gen_dep"`
	Inputs   []string `yaml:"inputs"`
	Outputs  []string `yaml:"outputs"`
}

type SumConfig struct {
	InputsHash  string `yaml:"inputs_hash"`
	OutputsHash string `yaml:"outputs_hash"`
}

type GenInfo struct {
	FilePath string
	GenCmd   string
}
