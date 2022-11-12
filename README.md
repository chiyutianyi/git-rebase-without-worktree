# git-rebase-without-worktree

This is a git-rebase-without-worktree demo.

With `git merge-tree --write-tree`, we can implement `git rebase` by merging each commit.

GIT_ALTERNATE_OBJECT_DIRECTORIES=$(PWD)/objects GIT_OBJECT_DIRECTORY=$(PWD)/objects/incoming-xxxxxx git-rebase-without-worktree <upstream> <branch>