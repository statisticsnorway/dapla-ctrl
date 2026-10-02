package iac

import (
	"context"
	"fmt"

	"github.com/google/go-github/v92/github"
)

const (
	initBranch              = "init"
	initialFilesTemplateDir = "templates/01-initial-files"
	teamRepoTemplateDir     = "templates/02-iac-repo-structure"
)

func initializeRepoContent(ctx context.Context, git *github.GitService, repoOwner, repoName, teamSlug string, isManaged bool) error {
	defaultBranch := "main"
	tmplVars := newTemplateVars(repoName, teamSlug, isManaged)

	initialFiles, err := renderTemplateDir(initialFilesTemplateDir, tmplVars)
	if err != nil {
		return fmt.Errorf("render initial files: %w", err)
	}
	repoStructureFiles, err := renderTemplateDir(teamRepoTemplateDir, tmplVars)
	if err != nil {
		return fmt.Errorf("render iac repo structure: %w", err)
	}

	if err := commitFiles(ctx, git, repoOwner, repoName, defaultBranch, defaultBranch, "Add initial files", initialFiles); err != nil {
		return fmt.Errorf("commit initial files to %s: %w", defaultBranch, err)
	}

	if err := commitFiles(ctx, git, repoOwner, repoName, defaultBranch, initBranch, "Add IaC repo structure", repoStructureFiles); err != nil {
		return fmt.Errorf("commit iac repo structure to %s: %w", initBranch, err)
	}

	return nil
}

func commitFiles(ctx context.Context, git *github.GitService, owner, repoName, baseBranch, branch, message string, files []repoFile) error {
	if len(files) == 0 {
		return nil
	}

	baseRef, _, err := git.GetRef(ctx, owner, repoName, "heads/"+baseBranch)
	if err != nil {
		return fmt.Errorf("get ref of branch %s: %w", baseBranch, err)
	}
	parentSHA := baseRef.GetObject().GetSHA()

	parent, _, err := git.GetCommit(ctx, owner, repoName, parentSHA)
	if err != nil {
		return fmt.Errorf("get commit %s: %w", parentSHA, err)
	}

	entries := make([]*github.TreeEntry, 0, len(files))
	for _, f := range files {
		entries = append(entries, &github.TreeEntry{
			Path:    new(f.Path),
			Mode:    new("100644"),
			Type:    new("blob"),
			Content: new(f.Content),
		})
	}
	tree, _, err := git.CreateTree(ctx, owner, repoName, parent.GetTree().GetSHA(), entries)
	if err != nil {
		return fmt.Errorf("create tree: %w", err)
	}

	commit, _, err := git.CreateCommit(ctx, owner, repoName, github.Commit{
		Message: &message,
		Tree:    tree,
		Parents: []*github.Commit{{SHA: &parentSHA}},
	}, nil)
	if err != nil {
		return fmt.Errorf("create commit: %w", err)
	}

	if branch == baseBranch {
		_, _, err = git.UpdateRef(ctx, owner, repoName, "heads/"+branch, github.UpdateRef{SHA: commit.GetSHA()})
	} else {
		_, _, err = git.CreateRef(ctx, owner, repoName, github.CreateRef{Ref: "refs/heads/" + branch, SHA: commit.GetSHA()})
	}
	if err != nil {
		return fmt.Errorf("point branch %s at commit %s: %w", branch, commit.GetSHA(), err)
	}
	return nil
}
