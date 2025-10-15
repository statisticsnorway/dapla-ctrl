.PHONY: default
default: | help

.PHONY: help
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST)\
	|awk -F ':[[:space:]]*## ' '{split($$1, a, ":"); printf "%-30s %s\n", a[2], $$2}'

.PHONY: build
build: ## Build the app
	pnpm install

.PHONY: run-dev
run-dev: ## Run the app in dev mode
	pnpm run dev

.PHONY: bump-version-patch
bump-version-patch: ## Bump patch version, e.g. 0.0.1 -> 0.0.2
	pnpm version patch

.PHONY: bump-version-minor
bump-version-minor: ## Bump minor version, e.g. 0.0.1 -> 0.1.0
	pnpm version minor

.PHONY: bump-version-major
bump-version-major: ## Bump major version, e.g. 0.0.1 -> 1.0.0
	pnpm version major

SHELL?=bash

.PHONY: shell
develop: ## Start the nix development shell
	nix develop -c $(SHELL)

.PHONY: build-local-docker
build-docker-local: ## Build the docker container
	docker build --platform linux/amd64 -t dapla-ctrl:latest .

include .env.local
.PHONY: run-docker-local
run-docker-local: ## Run the docker container
	./bin/run-docker.sh
