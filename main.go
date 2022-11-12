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

	// get commits to rebase by git rev-list
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
		if len(commit.parentIDs) > 1 {
			// this is commit commit
			/*
				-r, --rebase-merges[=(rebase-cousins|no-rebase-cousins)]
				By default, a rebase will simply drop merge commits from the todo list, and put the rebased commits into a single, linear branch. With --rebase-merges, the rebase will
				instead try to preserve the branching structure within the commits that are to be rebased, by recreating the merge commits. Any resolved merge conflicts or manual amendments
				in these merge commits will have to be resolved/re-applied manually.

				By default, or when no-rebase-cousins was specified, commits which do not have <upstream> as direct ancestor will keep their original branch point, i.e. commits that would be
				excluded by git-log(1)'s --ancestry-path option will keep their original ancestry by default. If the rebase-cousins mode is turned on, such commits are instead rebased onto
				<upstream> (or <onto>, if specified).

				It is currently only possible to recreate the merge commits using the ort merge strategy; different merge strategies can be used only via explicit exec git merge -s
				<strategy> [...] commands.

				See also REBASING MERGES and INCOMPATIBLE OPTIONS below.
			*/
			continue
		}
		log.Infof("[merge] merge commit %v %v", upstream, commitID)
		c := exec.Command("git", "merge-tree", "--write-tree", upstream, commitID)
		c.Env = os.Environ()

		out, err := c.CombinedOutput()
		if err != nil {
			log.Fatalf("git merge-tree failed: %v %v", string(out), err)
		}

		// get tree id from the result of merge-tree
		treeID := strings.Trim(string(out), "\n")

		c = exec.Command("git", "commit-tree", treeID, "-p", parent, "-m", commit.body)
		// use original author and committer
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
		// use current merge result for next parent
		parent = strings.Trim(string(out), "\n")
		log.Infof("[merge] merge result %v", parent)
	}
	fmt.Fprint(os.Stdout, parent)
}

type GitCommit struct {
	parentIDs      []string
	author         string
	authorEmail    string
	authorDate     string
	committer      string
	committerEmail string
	body           string
}

func getCommit(commitID string) *GitCommit {
	c := exec.Command("git", "show", "--pretty=format:%P%n%an%n%ae%n%ai%n%cn%n%ce%n%B", "-s", commitID)
	c.Env = os.Environ()

	out, err := c.CombinedOutput()
	if err != nil {
		log.Fatalf("git show failed: %v %v", string(out), err)
	}
	outs := strings.SplitN(string(out), "\n", 7)
	return &GitCommit{
		parentIDs:      strings.Split(outs[0], " "),
		author:         outs[1],
		authorEmail:    outs[2],
		authorDate:     outs[3],
		committer:      outs[4],
		committerEmail: outs[5],
		body:           outs[6],
	}
}
