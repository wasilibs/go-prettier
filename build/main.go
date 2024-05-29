package main

import (
	"github.com/goyek/x/boot"

	"github.com/wasilibs/tools/tasks"
)

func main() {
	tasks.Define(tasks.Params{
		LibraryName: "prettier",
		LibraryRepo: "prettier/prettier",
		GoReleaser:  true,
	})
	boot.Main()
}
