SERVICE_NAME=cms
GIT_COMMIT=`git rev-parse --short HEAD`
-include .env
export

.PHONY: setup
setup: ## Get linting stuffs
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	go get golang.org/x/tools/cmd/goimports

.PHONY: build-images
build-images: ## Build the images
	nerdctl build --platform=amd64,arm64 --tag containers.chewedfeed.com/chewedfeed/${SERVICE_NAME}:${GIT_COMMIT} --build-arg VERSION=0.1 --build-arg BUILD=${GIT_COMMIT} --build-arg SERVICE_NAME=${SERVICE_NAME} -f ./k8s/Containerfile .
	nerdctl tag containers.chewedfeed.com/chewedfed/${SERVICE_NAME}:${GIT_COMMIT} containers.chewedfeed.com/chewedfeed/${SERVICE_NAME}:latest

.PHONY: publish-images
publish-images:
	nerdctl push containers.chewedfeed.com/chewedfeed/${SERVICE_NAME}:${GIT_COMMIT} --all-platforms
	nerdctl push containers.chewedfeed.com/chewedfeed/${SERVICE_NAME}:latest --all-platforms

.PHONY: build
build: build-images

.PHONY: deploy
deploy:
	kubectl set image deployment/${SERVICE_NAME} ${SERVICE_NAME}=containers.chewedfeed.com/chewedfeed/${SERVICE_NAME}:${GIT_COMMIT} --namespace=retro-board

.PHONY: build-deploy
build-deploy: build publish-images deploy

.PHONY: lint-build-deploy
lint-build-deploy: lint build publish-images deploy

.PHONY: test
test: lint ## Test the app
	go test \
		-v \
		-race \
		-bench=./... \
		-benchmem \
		-timeout=120s \
		-cover \
		-coverprofile=./test/coverage.txt \
		-bench=./... ./...

.PHONY: mocks
mocks: ## Generate the mocks
	go generate ./...

.PHONY: full
full: clean build fmt lint test ## Clean, build, make sure its formatted, linted, and test it

.PHONY: lint
lint: ## Lint
	golangci-lint run --config configs/golangci.yml

.PHONY: fmt
fmt: ## Formatting
	gofmt -w -s .
	goimports -w .
	go clean ./...

.PHONY: pre-commit
pre-commit: fmt lint ## Do formatting and linting

.PHONY: clean
clean: ## Clean
	go clean ./...
	rm -rf bin/${SERVICE_NAME}
