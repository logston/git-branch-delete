# `git-branch-delete`

`git-branch-delete` deletes branches with no diff between them and a reference
branch. For example, if you are on your `main` branch, and you have local feature
branches that were merged remotely (eg. via GitHub), you can fast-forward your
local `main` branch and then run `git-branch-delete` to prune the merged
branches. Branches with diffs will not be pruned. Branches that can not be
rebased on the updated `main` branch will not be pruned.

## Installation

```
go get github.com/logston/git-branch-delete/settings
```

## Usage

```
➜  myrepo git:(main) git-branch-delete
Rebasing branches onto: main
Checking 'add-readme' ... branch has content. Moving on.
Checking 'go-1.16' ... branch has no content. Deleting...done.
Checking 'test' ... unable to rebase branch. Moving on.
```
