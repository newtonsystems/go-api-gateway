#
# Makefile
#

CURRENT_RELEASE_VERSION=0.0.1
REPO=go-api-gateway
LOCAL_DEPLOYMENT_FILENAME=api-deployment.yml
GO_MAIN=./app/cmd/addsvc/main.go
GO_PORT=50000


#
# Help for Makefile & Colorised Messages
#
# Powered by https://gist.github.com/prwhite/8168133
GREEN  := $(shell tput -Txterm setaf 2)
RED    := $(shell tput -Txterm setaf 1)
BLUE   := $(shell tput -Txterm setaf 4)
WHITE  := $(shell tput -Txterm setaf 7)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

INFO=$(GREEN)[INFO] $(RESET)
STAGE=$(BLUE)[INFO] $(RESET)
ERROR=$(RED)[ERROR] $(RESET)
WARN=$(YELLOW)[WARN] $(RESET)

#
# Help Command
#

# Add help text after each target name starting with '\#\#'
# A category can be added with @category
HELP_FUN = \
    %help; \
    while(<>) { push @{$$help{$$2 // 'options'}}, [$$1, $$3] if /^([a-zA-Z\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
    print "usage: make [target]\n\n"; \
    for (sort keys %help) { \
    print "${WHITE}$$_:${RESET}\n"; \
    for (@{$$help{$$_}}) { \
    $$sep = " " x (32 - length $$_->[0]); \
    print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
    }; \
    print "\n"; }

.PHONY: help

help:         ##@other Show this help.
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

# ------------------------------------------------------------------------------
.PHONY: update master build

update:       ##@build Updates dependencies for your go application
	bash -c "mkubectl.sh --update-deps"

install:      ##@build Install dependencies for your go application
	bash -c "mkubectl.sh --install-deps"

build:        ##@compile Builds executable cross compiled for alpine docker
	bash -c "mkubectl.sh --compile-inside-docker ${REPO} ${GO_MAIN}"

# ------------------------------------------------------------------------------
# CircleCI support
.PHONY: check

check:        ##@circleci Needed for running circleci tests
	@echo "$(INFO) Running tests"
	go test -v .

# ------------------------------------------------------------------------------
# Non docker local development (can be useful for super fast local/debugging)
.PHONY: run-conn-local

run-conn:    ##@dev Run locally (outside docker) (but connect to minikube linkerd etc)
	go run ${GO_MAIN} --conn.local

# ------------------------------------------------------------------------------
# Minikube (Normal Development)
.PHONY: run swap-hot-local swap-latest swap-latest-release

run:                    ##@dev Alias for swap-hot-local
	@make swap-hot-local

swap-hot-local:         ##@dev Swaps $(REPO) deployment in minikube with hot-reloadable docker image (You must make sure you are running i.e. infra-minikube.sh --create)
	@bash -c "mkubectl.sh --hot-reload-deployment ${REPO} ${LOCAL_DEPLOYMENT_FILENAME} ${GO_PORT}"

swap-latest:            ##@dev Swaps $(REPO) deployment in minikube with the latest image for branch from dockerhub (You must make sure you are running i.e. infra-minikube.sh --create)
	@bash -c "mkubectl.sh --swap-deployment-with-latest-image ${REPO} ${LOCAL_DEPLOYMENT_FILENAME}"

swap-latest-release:    ##@dev Swaps $(REPO) deployment in minikube with the latest release image for from dockerhub (You must make sure you are running i.e. infra-minikube.sh --create)
	@bash -c "mkubectl.sh --swap-deployment-with-latest-release-image ${REPO} ${LOCAL_DEPLOYMENT_FILENAME} ${CURRENT_RELEASE_VERSION}"
