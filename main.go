package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

var headReFormat = regexp.MustCompile(`ref: refs/heads/(?P<branch>.*)`)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {
	wd, err := os.Getwd()
	if err != nil {
		slog.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
	slog.Debug(fmt.Sprintf("wd: %s", wd))

	file, err := os.Open(path.Join(wd, ".git", "HEAD"))
	if errors.Is(err, os.ErrNotExist) {
		slog.Error("Not at root of repository.")
		os.Exit(1)
	}
	slog.Debug(fmt.Sprintf("file: %v", file))

	data := make([]byte, 1024)
	_, err = file.Read(data)
	if errors.Is(err, os.ErrNotExist) {
		slog.Error("Unable to read current branch.")
		os.Exit(1)
	}

	baseBranch, err := parseRefForBranch(string(data))
	if err != nil {
		slog.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("Rebasing branches onto: %s", baseBranch))

	branches, err := listBranches()
	if err != nil {
		slog.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}

	// checkout each branch
	for _, branch := range branches {
		slog.Debug(fmt.Sprintf("branch name: %s", branch))

		if branch == baseBranch {
			continue
		}

		slog.Info(fmt.Sprintf("Checking '%s' ... ", branch))

		if err := checkoutBranch(branch); err != nil {
			slog.Error(fmt.Sprintf("could not checkout branch %s: %v", branch, err))
			os.Exit(1)
		}

		success, err := rebaseBranch(baseBranch, branch)
		if err != nil {
			slog.Error(fmt.Sprintf("could not rebase branch %s: %v", branch, err))
			os.Exit(1)
		}

		if success {
			if err := checkoutBranch(baseBranch); err != nil {
				slog.Error(fmt.Sprintf("could not checkout base branch %s: %v", baseBranch, err))
				os.Exit(1)
			}

			hasContent, err := diffBranch(baseBranch, branch)
			if err != nil {
				slog.Error(fmt.Sprintf("could not diff branch %s: %v", branch, err))
				os.Exit(1)
			}

			if hasContent {
				fmt.Printf("branch has content. Moving on.\n")
			} else {
				fmt.Printf("branch has no content. Deleting...")
				if err := deleteBranch(branch); err != nil {
					slog.Error(fmt.Sprintf("could not delete branch %s: %v", branch, err))
					os.Exit(1)
				}
			}
		} else {
			fmt.Printf("unable to rebase branch. Moving on.\n")
		}
	}

	if err := checkoutBranch(baseBranch); err != nil {
		slog.Error(fmt.Sprintf("%v", err))
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

func listBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format", "%(refname:short)")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	branches := []string{}
	for _, b := range strings.Split(string(b), "\n") {
		b = strings.Trim(b, "'")
		branches = append(branches, b)
	}
	return branches, err
}

func rebaseBranch(baseBranch, branch string) (bool, error) {
	cmd := exec.Command("git", "rebase", baseBranch)

	// There are no platform independent ways to determine exit code. Thus we
	// use a hack to test if branch failed to rebase...
	stdOE, err := cmd.CombinedOutput()
	// First check if error was due to a rebase conflict.
	if strings.Contains(string(stdOE), "CONFLICT") && strings.Contains(string(stdOE), "abort") {
		if err := exec.Command("git", "rebase", "--abort").Run(); err != nil {
			return false, err
		}
		return false, nil
	}
	// If a non-rebase conflict error found, return error.
	if err != nil {
		return false, err
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
	if err := exec.Command("git", "branch", "-D", branch).Run(); err != nil {
		return err
	}
	slog.Info("done.")
	return nil
}
