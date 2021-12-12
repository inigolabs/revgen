package revgen

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortConfig(t *testing.T) {
	type testcase struct {
		in     Config
		expect Config
	}
	tests := []testcase{
		{
			in:     config(gen("a", "1"), gen("b", "1")),
			expect: config(gen("a", "1"), gen("b", "1")),
		},
		{
			in:     config(gen("b", "1"), gen("a", "1")),
			expect: config(gen("a", "1"), gen("b", "1")),
		},
		{
			in:     config(gen("a", "2"), gen("a", "1")),
			expect: config(gen("a", "1"), gen("a", "2")),
		},
		{
			in:     config(gen("a", "1"), gen("a", "2")),
			expect: config(gen("a", "1"), gen("a", "2")),
		},
	}

	for _, test := range tests {
		sort.Sort(test.in)
		require.Equal(t, test.in, test.expect)
	}
}

func TestSortStatus(t *testing.T) {
	type testcase struct {
		in     Status
		expect Status
	}
	tests := []testcase{
		{
			in:     status(sum("a", "1"), sum("b", "1")),
			expect: status(sum("a", "1"), sum("b", "1")),
		},
		{
			in:     status(sum("b", "1"), sum("a", "1")),
			expect: status(sum("a", "1"), sum("b", "1")),
		},
		{
			in:     status(sum("a", "2"), sum("a", "1")),
			expect: status(sum("a", "1"), sum("a", "2")),
		},
		{
			in:     status(sum("a", "1"), sum("a", "2")),
			expect: status(sum("a", "1"), sum("a", "2")),
		},
	}

	for _, test := range tests {
		sort.Sort(test.in)
		require.Equal(t, test.in, test.expect)
	}
}

func gen(filepath string, gencmd string) *GenConfig {
	return &GenConfig{
		FilePath: filepath,
		GenCmd:   gencmd,
	}
}

func config(configs ...*GenConfig) Config {
	return Config{
		Configs: configs,
	}
}

func sum(filepath string, gencmd string) *SumConfig {
	return &SumConfig{
		FilePath: filepath,
		GenCmd:   gencmd,
	}
}

func status(configs ...*SumConfig) Status {
	return configs
}
