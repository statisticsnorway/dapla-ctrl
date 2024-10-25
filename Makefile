.PHONY: default
default: | help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the app
	npm install

build-docker-local:
	docker build -t dapla-ctrl .

.PHONY: run-dev 
run-dev: ## Run the app in dev mode
	npm run dev

include .env.local
run-docker-local:
	docker run -it -p 8080:8080 -e VITE_JWKS_URI=${VITE_JWKS_URI} dapla-ctrl