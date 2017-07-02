#
# Makefile
#
# Created originally by cookiecutter: version X
#

REPO=go-hello
PROJECT_NAME=hello

TIMESTAMP=tmp-$(shell date +%s )

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

INFO=$(GREEN)====>>[INFO]$(RESET)
ERROR=$(RED)====>>[ERROR]$(RESET)
WARN=$(YELLOW)====>>[WARN]$(RESET)

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



#
# Other
#
.PHONY: help help-show-normal-usage help-how-to

help:                        ##@other Show this help.
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

help-show-normal-usage:      ##@other Shows normal usage case for development (what commands to run)
	@echo "${GREEN}The normal usage is the following for working locally:${RESET}"
	@echo "\tIf you want to run the services locally there are three ways:"
	@echo "\t\t1. Develop & Deploy to minikube"
	@echo "\t\t2. Run docker container locally (using minikube docker environment) + other services in minikube"
	@echo "\t\t3. Run docker container locally + nghttpx + other services in minikube (Recommended)"
	@echo ""
	@echo ""
	@echo "1. Develop & Deploy to minikube"
	@echo "${YELLOW}NOTE: Current docker-machine doesn't support inotify - therefore no hot-reloading, doh!${RESET}"
	@echo ""
	@echo "A normal command workflow may be:"
	@echo "\t${GREEN}make infra-create${RESET}"
	@echo "\t${GREEN}make kube-create${RESET}"
	@echo "\t## Make some dev changes ##"
	@echo "\t${GREEN}make kube-update${RESET}"
	@echo ""
	@echo ""
	@echo "2. Run docker container locally (using minikube docker environment) + other services in minikube"
	@echo ""
	@echo "\t${GREEN}make infra-create${RESET}"
	@echo "\t${GREEN}make run-dm${RESET}"
	@echo "\t## (You must use incoming-dev-workflow-1 to connect external to services) ##"
	@echo "\t## ctrl+c (stop running container) then make some dev changes ##"
	@echo "\t## Maybe a make build-dm ##"
	@echo "\t${GREEN}make run-dm${RESET}"
	@echo ""
	@echo ""
	@echo "3. Run docker container locally + nghttpx + other services in minikube (Recommended)"
	@echo "${YELLOW}NOTE: We must use a special router in linkerd to statically router to a nghttpx reverse proxy (which points to a local running docker container)${RESET}"
	@echo ""
	@echo "\t${GREEN}make infra-create${RESET}"
	@echo "\t${GREEN}make nghttpx${RESET}"
	@echo "\t${GREEN}make run${RESET}"
	@echo "\t## (You must use incoming-dev-workflow-2 to connect external to services) ##"
	@echo "\t## ctrl+c (stop running container) then make some dev changes ##"
	@echo "\t## Maybe a make build ##"
	@echo "\t${GREEN}make run${RESET}"
	@echo ""
	@echo ""

help-how-to:                 ##@other Shows some useful answers to frequent questions
	@echo "$(GREEN)Questions & Answers - How to guide$(RESET)"
	@echo ""
	@echo "$(GREEN)How to add a local python package:$(RESET)"
	@echo "\tYou can add a local python package easily by adding a volume to DOCKER_RUN_COMMAND / DOCKER_RUN_LOCAL_COMMAND"
	@echo "\tObviously make sure it is in the requirements - which it should be already"
	@echo ""
	@echo "$(GREEN)How to ignore the Prerequisites for running this makefile:$(RESET)"
	@echo "\tYou can ignore the Prerequisites by setting FORCE_IGNORE_PREQ_TEST: make <command> FORCE_IGNORE_PREQ_TEST=true"
	@echo ""

export WORKDIR=/Users/danvir/go/src/go-hello/app

#
# Main
#
build-go:
	@echo "$(INFO) Getting packages and building alpine go binary ..."
	go get ./...
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o main ./app 

build:                       ##@local Builds the local Dockerfile
	@echo "$(INFO) Building a linux-alpine Go binary locally"
	docker run --rm -v "${PWD}/app":${WORKDIR} -w ${WORKDIR} -f Dockerfile.build .
	#cd app/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main
	@echo ""
	#chmod +x app/main
	@echo "$(INFO) Building docker container locally with tag: $(BLUE)$(REPO):local$(RESET)"
	#docker image build -t $(REPO):local -f Dockerfile .
	@echo ""

run: build                   ##@local Builds and run docker container with tag: '$(REPO):local' as a one-off. ##@local (dev-workflow-2) Runs docker container on same network as minikube making it accessible from kubernetes minikube and other kubernetes services
	@echo "$(INFO) Running docker container with tag: $(REPO):local"
	@echo "$(BLUE)"
	docker run $(REPO):local
	@echo "$(NO_COLOR)"

