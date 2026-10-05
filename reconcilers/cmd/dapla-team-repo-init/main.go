// One-off command to init dapla team repositories that because of some reason did not
// get created during first reconcile.
// We should than run this script and commit and create PR (init branch).
// Usage:
//
//	go run ./cmd/dapla-team-repo-init -team my-team -managed=true [-repo my-team-iac]
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/github/iac"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	teamSlug := flag.String("team", "", "slug of the team owning the repo (required)")
	repoName := flag.String("repo", "", `name of the repo to initialize (default "<team>-iac")`)
	isManaged := flag.Bool("managed", false, "whether the team is managed (required to be set explicitly, e.g. -managed=false)")
	outputPath := flag.String("output", "", `path of where to output the files (default "./<repo>/")`)
	flag.Parse()

	managedSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "managed" {
			managedSet = true
		}
	})

	if *teamSlug == "" {
		flag.Usage()
		return errors.New("-team is required")
	}
	if !managedSet {
		flag.Usage()
		return errors.New("-managed must be set explicitly")
	}
	if *repoName == "" {
		*repoName = *teamSlug + "-iac"
	}
	if *outputPath == "" {
		*outputPath = "./" + *repoName + "/"
	}

	log.Printf("initializing dapla team repo %s for %s (managed=%t). Writing to path %s\n", *repoName, *teamSlug, *isManaged, *outputPath)

	vars := iac.NewTemplateVars(*repoName, *teamSlug, *isManaged)
	initFiles, err := iac.RenderTemplateDir(iac.InitialFilesTemplateDir, vars)
	if err != nil {
		return err
	}
	teamFiles, err := iac.RenderTemplateDir(iac.TeamRepoTemplateDir, vars)
	if err != nil {
		return err
	}

	for _, f := range initFiles {
		writeFile(*outputPath+f.Path, f.Content)
	}
	for _, f := range teamFiles {
		writeFile(*outputPath+f.Path, f.Content)
	}

	log.Println("done! Move the files to the team iac-repo. Checkout a new branch named 'init', commit and create a PR, which atlantis will run.")
	return nil
}

func writeFile(path, content string) {
	dirname, _, _ := strings.CutLast(path, "/")
	err := os.MkdirAll(dirname, 0o755)
	if err != nil {
		log.Fatalf("failed to create directory: %v", err)
	}

	// 3. Write data to the file
	// 0644 gives the owner read/write permissions, and others read-only
	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		log.Fatalf("failed to write file: %v", err)
	}
	log.Println("wrote", path)
}
