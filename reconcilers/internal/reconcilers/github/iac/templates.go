package iac

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

const (
	templateSuffix = ".tmpl"

	managed     = "MANAGED"
	selfManaged = "SELF_MANAGED"
)

//go:embed all:templates
var repoTemplates embed.FS

type templateVariables struct {
	RepositoryName string
	TeamSlug       string
	AutonomyLevel  string
}

func NewTemplateVars(repoName string, teamSlug string, isManaged bool) templateVariables {
	autonomyLevel := managed
	if !isManaged {
		autonomyLevel = selfManaged
	}
	return templateVariables{
		RepositoryName: repoName,
		TeamSlug:       teamSlug,
		AutonomyLevel:  autonomyLevel,
	}
}

type repoFile struct {
	Path    string
	Content string
}

func RenderTemplateDir(dir string, tmplVars templateVariables) ([]repoFile, error) {
	var files []repoFile
	err := fs.WalkDir(repoTemplates, dir, func(templatePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		relativePath := strings.TrimPrefix(templatePath, dir+"/")

		if tmplVars.AutonomyLevel == selfManaged && strings.HasPrefix(relativePath, "automation/") {
			// Only managed teams should have automation folder
			return nil
		}

		outputPath, err := getValidatedOutputPath(relativePath, tmplVars, templateSuffix)
		if err != nil {
			return err
		}

		content, err := renderFile(templatePath, tmplVars)
		if err != nil {
			return err
		}

		files = append(files, repoFile{Path: outputPath, Content: content})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func getValidatedOutputPath(relativePath string, tmplVars templateVariables, templateSuffix string) (string, error) {
	path := relativePath

	// assume path should be templated if brackets are present
	// e.g. automation which has the team slug in directory name
	if strings.Contains(relativePath, "{{") {
		var err error
		path, err = executeTemplate(relativePath, tmplVars)
		if err != nil {
			return "", fmt.Errorf("render path of template %q: %w", relativePath, err)
		}
	}

	if !fs.ValidPath(path) || path == "." {
		return "", fmt.Errorf("template %q rendered to invalid path %q", relativePath, path)
	}
	return strings.TrimSuffix(path, templateSuffix), nil
}

func renderFile(templatePath string, tmplVars templateVariables) (string, error) {
	contents, err := repoTemplates.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read template %q: %w", templatePath, err)
	}
	content := string(contents)
	if strings.HasSuffix(templatePath, templateSuffix) {
		// only template if the file actually is a template file
		content, err = executeTemplate(content, tmplVars)
		if err != nil {
			return "", fmt.Errorf("render template %q: %w", templatePath, err)
		}
	}
	return content, nil
}

func executeTemplate(source string, data templateVariables) (string, error) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(source)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", err
	}
	return rendered.String(), nil
}
