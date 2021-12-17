package revgen

type Config struct {
	AutoUpdate bool         `yaml:"auto_update"`
	Configs    []*GenConfig `yaml:"configs"`
}

type Status []*SumConfig

type Key struct {
	FilePath string `yaml:"path"`
	GenCmd   string `yaml:"cmd"`
}

type GenConfig struct {
	FilePath string   `yaml:"path"`
	GenCmd   string   `yaml:"gen_cmd"`
	GenDeps  []string `yaml:"gen_deps"`
	GenFiles []string `yaml:"gen_files"`
}

type SumConfig struct {
	FilePath  string `yaml:"path"`
	GenCmd    string `yaml:"cmd"`
	HashDeps  string `yaml:"hash_deps"`
	HashFiles string `yaml:"hash_files"`
}

type ConfigMap struct {
	AutoUpdate bool
	Configs    map[Key]*GenConfig
}

type StatusMap = map[Key]*SumConfig

func (l Config) Len() int {
	return len(l.Configs)
}

func (l Config) Less(i int, j int) bool {
	if l.Configs[i].FilePath < l.Configs[j].FilePath {
		return true
	} else if l.Configs[i].FilePath == l.Configs[j].FilePath {
		if l.Configs[i].GenCmd < l.Configs[j].GenCmd {
			return true
		}
	}
	return false
}

func (l Config) Swap(i int, j int) {
	buf := l.Configs[i]
	l.Configs[i] = l.Configs[j]
	l.Configs[j] = buf
}

func (l Status) Len() int {
	return len(l)
}

func (l Status) Less(i int, j int) bool {
	if l[i].FilePath < l[j].FilePath {
		return true
	} else if l[i].FilePath == l[j].FilePath {
		if l[i].GenCmd < l[j].GenCmd {
			return true
		}
	}
	return false
}

func (l Status) Swap(i int, j int) {
	buf := l[i]
	l[i] = l[j]
	l[j] = buf
}
