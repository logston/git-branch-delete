package main

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Running for %s...\n", wd)

	file, err := os.Open(path.Join(wd, ".git", "HEAD"))
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Not at root of repository.")
		os.Exit(1)
	}

	fmt.Println(file)

	// list branches

	// checkout each branch

	// rebase on master

	// If any issues, abort
	// If no issues,
	//    If nothing there, great delete branch
	//    If soemthing there, leave it be.

	fmt.Println("vim-go")
}
