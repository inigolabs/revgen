package revgen

type Key struct {
	FilePath string `yaml:"path"`
	GenCmd   string `yaml:"cmd"`
}

type GenConfig struct {
	FilePath string   `yaml:"path"`
	GenCmd   string   `yaml:"cmd"`
	GenDeps  []string `yaml:"deps"`
}

type GenMap = map[Key]*GenConfig
type GenList struct {
	Configs []*GenConfig `yaml:"configs"`
}

type SumConfig struct {
	FilePath string `yaml:"path"`
	GenCmd   string `yaml:"cmd"`
	Hash     string `yaml:"hash"`
}

type SumMap = map[Key]*SumConfig
type SumList []*SumConfig

func (l GenList) Len() int {
	return len(l.Configs)
}

func (l GenList) Less(i int, j int) bool {
	if l.Configs[i].FilePath < l.Configs[j].FilePath {
		return true
	} else if l.Configs[i].FilePath == l.Configs[j].FilePath {
		if l.Configs[i].GenCmd < l.Configs[j].GenCmd {
			return true
		}
	}
	return false
}

func (l GenList) Swap(i int, j int) {
	buf := l.Configs[i]
	l.Configs[i] = l.Configs[j]
	l.Configs[j] = buf
}

func (l SumList) Len() int {
	return len(l)
}

func (l SumList) Less(i int, j int) bool {
	if l[i].FilePath < l[j].FilePath {
		return true
	} else if l[i].FilePath == l[j].FilePath {
		if l[i].GenCmd < l[j].GenCmd {
			return true
		}
	}
	return false
}

func (l SumList) Swap(i int, j int) {
	buf := l[i]
	l[i] = l[j]
	l[j] = buf
}
