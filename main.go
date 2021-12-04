// +build ignore

package main

import (
	"os"

	"github.com/inigolabs/revgen"
)

func main() {
	revgen.Generate(os.Args)
}
