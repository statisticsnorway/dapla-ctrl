package iac

import (
	"context"
	"fmt"

	"github.com/google/go-github/v92/github"
)

const (
	defaultBranch           = "main"
	initBranch              = "init"
	InitialFilesTemplateDir = "templates/01-initial-files"
	TeamRepoTemplateDir     = "templates/02-dapla-team-repo"
)

type repoContentService struct {
	git *github.GitService
}

func (g *repoContentService) initIacRepoContent(ctx context.Context, repoOwner, repoName, teamSlug string, isManaged bool) error {
	tmplVars := NewTemplateVars(repoName, teamSlug, isManaged)

	err := g.templateAndCommitFiles(ctx, tmplVars, repoOwner, repoName)
	if err != nil {
		return err
	}

	err = g.templateAndCommitFiles(ctx, tmplVars, repoOwner, repoName)
	if err != nil {
		return err
	}

	return nil
}

func (g *repoContentService) templateAndCommitFiles(ctx context.Context, tmplVars templateVariables, repoOwner string, repoName string) error {
	initialFiles, err := RenderTemplateDir(InitialFilesTemplateDir, tmplVars)
	if err != nil {
		return fmt.Errorf("render initial files: %w", err)
	}
	if err := g.commitFiles(ctx, repoOwner, repoName, defaultBranch, defaultBranch, "Add initial files", initialFiles); err != nil {
		return fmt.Errorf("commit initial files to %s: %w", defaultBranch, err)
	}
	return nil
}

func (g *repoContentService) commitFiles(ctx context.Context, owner, repoName, baseBranch, branch, message string, files []repoFile) error {
	if len(files) == 0 {
		return nil
	}

	baseRef, _, err := g.git.GetRef(ctx, owner, repoName, "heads/"+baseBranch)
	if err != nil {
		return fmt.Errorf("get ref of branch %s: %w", baseBranch, err)
	}
	parentSHA := baseRef.GetObject().GetSHA()

	parent, _, err := g.git.GetCommit(ctx, owner, repoName, parentSHA)
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
	tree, _, err := g.git.CreateTree(ctx, owner, repoName, parent.GetTree().GetSHA(), entries)
	if err != nil {
		return fmt.Errorf("create tree: %w", err)
	}

	commit, _, err := g.git.CreateCommit(ctx, owner, repoName, github.Commit{
		Message: &message,
		Tree:    tree,
		Parents: []*github.Commit{{SHA: &parentSHA}},
	}, nil)
	if err != nil {
		return fmt.Errorf("create commit: %w", err)
	}

	if branch == baseBranch {
		_, _, err = g.git.UpdateRef(ctx, owner, repoName, "heads/"+branch, github.UpdateRef{SHA: commit.GetSHA()})
	} else {
		_, _, err = g.git.CreateRef(ctx, owner, repoName, github.CreateRef{Ref: "refs/heads/" + branch, SHA: commit.GetSHA()})
	}
	if err != nil {
		return fmt.Errorf("point branch %s at commit %s: %w", branch, commit.GetSHA(), err)
	}
	return nil
}
