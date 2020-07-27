GOFMT_FILES?=$$(find . -not -path "./vendor/*" -type f -name '*.go')
PROJECT_NAME?=unpackker
APP_DIR?=$$(git rev-parse --show-toplevel)
VERSION?=0.1
DEV?=${DEVBOX_TRUE}

.PHONY: help
help: ## Prints help (only for targets with comments)
	@grep -E '^[a-zA-Z0-9._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

local.fmt: ## Lints all the go code in the application.
	gofmt -w $(GOFMT_FILES)

local.check: local.fmt ## Loads all the dependencies to vendor directory
	go mod vendor
	go mod tidy

local.build: local.check ## Generates the artifact with the help of 'go build'
	go build -o $(PROJECT_NAME) -ldflags="-s -w"

local.push: local.build ## Pushes built artifact to the specified location

local.run: local.build ## Generates the artifact and start the service in the current directory
	./${PROJECT_NAME}

dockerise: local.check ## Containerise the appliction
	docker build . --tag ${DOCKER_USER}/${PROJECT_NAME}:${VERSION}

docker.lint: ## Linting Dockerfile for
	if [ -z "${DEV}" ]; then hadolint Dockerfile ; else docker run --rm -v $(APP_DIR):/app -w /app hadolint/hadolint:latest-alpine hadolint Dockerfile ; fi

docker.login: ## Establishes the connection to the docker registry
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASSWD} ${DOCKER_REPO}

docker.publish.image: docker_login ## Publisies the image to the registered docker registry.
	docker push ${DOCKER_USER}/${PROJECT_NAME}:${VERSION}

coverage.lint: ## Lint's application for errors, it is a linters aggregator (https://github.com/golangci/golangci-lint).
	if [ -z "${DEV}" ]; then golangci-lint run --color always ; else docker run --rm -v $(APP_DIR):/app -w /app golangci/golangci-lint:v1.27-alpine golangci-lint run --color always ; fi

coverage.report: ## Publishes the go-report of the appliction (uses go-reportcard)
	if [ -z "${DEV}" ]; then goreportcard-cli -v ; else docker run --rm -v $(APP_DIR):/app -w /app basnik/goreportcard-cli:latest goreportcard-cli -v ; fi

dev.prerequisite.up: ## Sets up the development environment with all necessary components.
	$(APP_DIR)/scripts/prerequisite.sh

dev.prerequisite.purge: ## Teardown the development environment by removing all components.

install.hooks: ## install pre-push hooks for the repository.
	${APP_DIR}/scripts/hook.sh ${APP_DIR}
