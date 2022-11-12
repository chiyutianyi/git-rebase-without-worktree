package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		log.Fatalf("usage: %s <upstream> [<branch>\n", os.Args[0])
	}

	upstream := args[1]
	branch := args[2]

	c := exec.Command("git", "rev-list", fmt.Sprintf("%s..%s", upstream, branch))
	c.Env = os.Environ()

	out, err := c.CombinedOutput()
	if err != nil {
		log.Fatalf("git rev-list failed: %v, %v", string(out), err)
	}

	if len(out) == 0 {
		// nothing to rebase
		return
	}

	parent := upstream

	commitIDs := strings.Split(string(out), "\n")

	for i := len(commitIDs) - 1; i >= 0; i-- {
		commitID := commitIDs[i]
		if commitID == "" {
			continue
		}
		// range each commitID and merge
		commit := getCommit(commitID)
		c := exec.Command("git", "merge-tree", "--write-tree", parent, commitID)
		c.Env = os.Environ()

		out, err := c.CombinedOutput()
		if err != nil {
			log.Fatalf("git merge-tree failed: %v %v", string(out), err)
		}
		treeID := strings.Trim(string(out), "\n")
		c = exec.Command("git", "commit-tree", treeID, "-p", parent, "-m", commit.body)
		c.Env = append(
			os.Environ(),
			fmt.Sprintf("GIT_AUTHOR_NAME=%s", commit.author),
			fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", commit.authorEmail),
			fmt.Sprintf("GIT_AUTHOR_DATE=%s", commit.authorDate),
			fmt.Sprintf("GIT_COMMITTER_NAME=%s", commit.committer),
			fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", commit.committerEmail),
		)

		out, err = c.CombinedOutput()
		if err != nil {
			log.Fatalf("git commit-tree failed: %v %v", string(out), err)
		}
		parent = strings.Trim(string(out), "\n")
	}
	fmt.Fprint(os.Stdout, parent)
}

type GitCommit struct {
	author         string
	authorEmail    string
	authorDate     string
	committer      string
	committerEmail string
	body           string
}

func getCommit(commitID string) *GitCommit {
	c := exec.Command("git", "show", "--pretty=format:%an%n%ae%n%ai%n%cn%n%ce%n%B", "-s", commitID)
	c.Env = os.Environ()

	out, err := c.CombinedOutput()
	if err != nil {
		log.Fatalf("git show failed: %v %v", string(out), err)
	}
	outs := strings.SplitN(string(out), "\n", 6)
	return &GitCommit{
		author:         outs[0],
		authorEmail:    outs[1],
		authorDate:     outs[2],
		committer:      outs[3],
		committerEmail: outs[4],
		body:           outs[5],
	}
}
