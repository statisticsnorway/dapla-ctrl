package iac

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestRenderTemplateDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		isManaged         bool
		codeOwnersSnippet string
		automationExists  bool
	}{
		{
			name:              "managed",
			isManaged:         true,
			codeOwnersSnippet: "*.yaml @statisticsnorway/example-data-admins",
			automationExists:  true,
		},
		{
			name:              "self managed",
			isManaged:         false,
			codeOwnersSnippet: "* @statisticsnorway/example-developers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmplVars := newTemplateVars("example-iac", "example", tt.isManaged)

			initial := renderToMap(t, initialFilesTemplateDir, tmplVars)
			if _, ok := initial[".github/workflows/render-projects.yaml"]; !ok || len(initial) != 1 {
				t.Fatalf("unexpected initial files: %v", paths(initial))
			}

			repoFiles := renderToMap(t, teamRepoTemplateDir, tmplVars)
			for _, path := range []string{".gitignore", ".github/workflows/kildomaten.yaml", "infra/projects.yaml", "infra/projects/README.md", "README.md", "CODEOWNERS"} {
				if _, ok := repoFiles[path]; !ok {
					t.Errorf("expected %q in rendered files: %v", path, paths(repoFiles))
				}
			}
			t.Run("template in dir name should render", func(t *testing.T) {
				for path := range repoFiles {
					if strings.HasSuffix(path, templateSuffix) || strings.Contains(path, "{{") {
						t.Errorf("path %q was not rendered", path)
					}
				}
			})

			t.Run("files should be templated", func(t *testing.T) {
				if !strings.Contains(repoFiles["README.md"], "# example-iac") {
					t.Errorf("README.md did not contain the repository name:\n%s", repoFiles["README.md"])
				}
				if !strings.Contains(repoFiles["infra/projects.yaml"], "team_uniform_name: example") {
					t.Errorf("projects.yaml did not contain the team slug:\n%s", repoFiles["infra/projects.yaml"])
				}
				if !strings.Contains(repoFiles["CODEOWNERS"], tt.codeOwnersSnippet) {
					t.Errorf("CODEOWNERS did not contain %q:\n%s", tt.codeOwnersSnippet, repoFiles["CODEOWNERS"])
				}

				_, automationExists := repoFiles["automation/source-data/example-prod/README.md"]
				if automationExists != tt.automationExists {
					t.Errorf("automation README exists = %v, want %v", automationExists, tt.automationExists)
				}
			})
		})
	}
}

func TestExecuteTemplateMissingFieldShouldError(t *testing.T) {
	t.Parallel()
	if _, err := executeTemplate("{{ .DoesNotExist }}", templateVariables{}); err == nil {
		t.Fatal("expected error for unknown template field")
	}
}

func renderToMap(t *testing.T, dir string, data templateVariables) map[string]string {
	t.Helper()
	files, err := renderTemplateDir(dir, data)
	if err != nil {
		t.Fatalf("renderTemplateDir(%q) error = %v", dir, err)
	}
	m := make(map[string]string, len(files))
	for _, f := range files {
		m[f.Path] = f.Content
	}
	return m
}

func paths(m map[string]string) []string {
	return slices.Collect(maps.Keys(m))
}
