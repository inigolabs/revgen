package main

import (
	"os"

	revgen "github.com/inigolabs/revgen/pkg"
)

func main() {
	revgen.Generate(os.Args)
}
