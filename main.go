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

	fmt.Println("Rebasing branches onto:", currentBranch)

	files, err := os.ReadDir(path.Join(wd, ".git", "refs", "heads"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// checkout each branch
	for _, file := range files {
		branch := file.Name()

		if branch == currentBranch {
			continue
		}

		fmt.Printf("Checking %s", branch)

		if err := checkoutBranch(branch); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		success, err := rebaseBranch(currentBranch, branch)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if success {
			if err := checkoutBranch(currentBranch); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			hasContent, err := diffBranch(currentBranch, branch)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !hasContent {
				if err := deleteBranch(branch); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}

	if err := checkoutBranch(currentBranch); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
	return exec.Command("git", "checkout", branch).Run()
}

func rebaseBranch(baseBranch, branch string) (bool, error) {
	cmd := exec.Command("git", "rebase", baseBranch)

	// There are no platform independent ways to determine exit code. Thus we
	// use a hack to test if branch failed to rebase...
	stdOE, cmdErr := cmd.CombinedOutput()
	if strings.Contains(string(stdOE), "CONFLICT") && strings.Contains(string(stdOE), "abort") {
		fmt.Printf("Unable to rebase %s. Rolling back...\n", branch)
		if err := exec.Command("git", "rebase", "--abort").Run(); err != nil {
			return false, err
		}
	} else if cmdErr != nil {
		return false, cmdErr
	}

	return true, nil
}

func diffBranch(baseBranch, branch string) (bool, error) {
	cmd := exec.Command("git", "diff", baseBranch, branch)

	stdOE, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	if len(stdOE) != 0 {
		return true, nil
	}

	return false, nil
}

func deleteBranch(branch string) error {
	fmt.Println("deleted branch")
	return nil
}
