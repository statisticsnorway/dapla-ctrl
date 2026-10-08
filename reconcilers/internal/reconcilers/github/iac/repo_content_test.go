package iac

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v92/github"
)

type fakeGitClient struct {
	trees      [][]*github.TreeEntry
	commits    int
	updatedRef github.CreateRef
	createdRef github.CreateRef
}

func (f *fakeGitClient) GetRef(context.Context, string, string, string) (*github.Reference, *github.Response, error) {
	return &github.Reference{Object: &github.GitObject{SHA: new("base-sha")}}, nil, nil
}

func (f *fakeGitClient) GetCommit(_ context.Context, _, _, sha string) (*github.Commit, *github.Response, error) {
	return &github.Commit{SHA: &sha, Tree: &github.Tree{SHA: new("base-tree-sha")}}, nil, nil
}

func (f *fakeGitClient) CreateTree(_ context.Context, _, _, _ string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error) {
	f.trees = append(f.trees, entries)
	return &github.Tree{SHA: new(fmt.Sprintf("tree-%d", len(f.trees)))}, nil, nil
}

func (f *fakeGitClient) CreateCommit(context.Context, string, string, github.Commit, *github.CreateCommitOptions) (*github.Commit, *github.Response, error) {
	f.commits++
	return &github.Commit{SHA: new(fmt.Sprintf("commit-%d", f.commits))}, nil, nil
}

func (f *fakeGitClient) UpdateRef(_ context.Context, _, _, ref string, body github.UpdateRef) (*github.Reference, *github.Response, error) {
	f.updatedRef = github.CreateRef{
		Ref: ref,
		SHA: body.SHA,
	}
	return nil, nil, nil
}

func (f *fakeGitClient) CreateRef(_ context.Context, _, _ string, body github.CreateRef) (*github.Reference, *github.Response, error) {
	f.createdRef = body
	return nil, nil, nil
}

func TestInitIacRepoContent(t *testing.T) {
	t.Parallel()

	client := &fakeGitClient{}
	r := &repoContentReconciler{client: client}

	if err := r.initIacRepoContent(t.Context(), "ssb", "play-iac", "play", true); err != nil {
		t.Fatal(err)
	}

	if client.commits != 2 {
		t.Fatalf("commits = %d, want 2", client.commits)
	}
	for i, entries := range client.trees {
		if len(entries) == 0 {
			t.Errorf("tree %d has no entries", i)
		}
	}

	// check first commit om main
	if want := (github.CreateRef{Ref: "heads/" + defaultBranch, SHA: "commit-1"}); client.updatedRef != want {
		t.Errorf("updated %+v, expected %+v", client.updatedRef, want)
	}
	// check init branch
	if want := (github.CreateRef{Ref: "refs/heads/" + initBranch, SHA: "commit-2"}); client.createdRef != want {
		t.Errorf("created %+v, expected %+v", client.createdRef, want)
	}
}
