package revgen

import (
	"reflect"
	"testing"
)

func TestOrderConfigs(t *testing.T) {
	type testcase struct {
		name   string
		config *Config
		expect []Name
	}
	tests := []testcase{
		{
			name: "NoDeps",
			config: &Config{
				Generators: map[Name]*GenConfig{
					Name("A"): {},
					Name("B"): {},
					Name("C"): {},
				},
			},
			expect: []Name{"A", "B", "C"},
		},
		{
			name: "EmptyDeps",
			config: &Config{
				Generators: map[Name]*GenConfig{
					Name("A"): {
						GenDep: "",
					},
					Name("B"): {
						GenDep: "",
					},
					Name("C"): {
						GenDep: "",
					},
				},
			},
			expect: []Name{"A", "B", "C"},
		},
		{
			name: "Deps",
			config: &Config{
				Generators: map[Name]*GenConfig{
					Name("C"): {
						GenDep: Name("B"),
					},
					Name("B"): {
						GenDep: Name("A"),
					},
					Name("A"): {},
				},
			},
			expect: []Name{"A", "B", "C"},
		},
		{
			name: "Duel",
			config: &Config{
				Generators: map[Name]*GenConfig{
					Name("AA"): {
						GenDep: Name("A"),
					},
					Name("BB"): {
						GenDep: Name("B"),
					},
					Name("A"): {},
					Name("B"): {},
				},
			},
			expect: []Name{"A", "AA", "B", "BB"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := orderConfigs(test.config)
			same := reflect.DeepEqual(test.expect, actual)
			if !same {
				t.Fatalf("mismatch\n expect:%v\n actual:%v\n", test.expect, actual)
			}
		})
	}
}
