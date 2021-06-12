package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

var headReFormat = regexp.MustCompile(`ref: refs/heads/(?P<branch>.*)`)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	file, err := os.Open(path.Join(wd, ".git", "HEAD"))
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Not at root of repository.")
		os.Exit(1)
	}

	data := make([]byte, 1024)
	_, err = file.Read(data)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Unable to read current branch.")
		os.Exit(1)
	}

	currentBranch, err := parseRefForBranch(string(data))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Rebasing branches onto: ", currentBranch)

	files, err := os.ReadDir(path.Join(wd, ".git", "refs", "heads"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// checkout each branch
	for _, file := range files {
		branch := file.Name()
		fmt.Println(branch)
		if err := checkoutBranch(branch); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := rebaseBranch(currentBranch, branch); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := checkoutBranch(currentBranch); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// If any issues, abort
	// If no issues,
	//    If nothing there, great delete branch
	//    If soemthing there, leave it be.
}

func parseRefForBranch(ref string) (string, error) {
	match := headReFormat.FindStringSubmatch(ref)
	result := make(map[string]string)
	for i, name := range headReFormat.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	branch, ok := result["branch"]
	if !ok {
		return "", errors.New("Unable to parse branch name.")
	}

	return branch, nil
}

func checkoutBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	stdOE, err := cmd.CombinedOutput()
	fmt.Println(string(stdOE))
	if err != nil {
		return err
	}
	return nil
}

func rebaseBranch(baseBranch, branch string) error {
	cmd := exec.Command("git", "rebase", baseBranch)
	// test

	// This is it
	stdOE, _ := cmd.CombinedOutput()

	// There are no platform indepenent ways to determine exit code. Thus we
	// use a hack to test if branch failed to rebase...
	if strings.Contains(string(stdOE), "CONFLICT") && strings.Contains(string(stdOE), "abort") {
		fmt.Println(">>>", string(stdOE), "<<<")
		fmt.Printf("Unable to rebase %s. Rolling back...\n", branch)
		if err := exec.Command("git", "rebase", "--abort").Run(); err != nil {
			return err
		}
	}

	return nil
}
