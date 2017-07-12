#
# Makefile
#
# Created originally by cookiecutter: version X
#

REPO=go-api-gateway
PROJECT_NAME=apt-gateway
CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
CURRENT_RELEASE_VERSION=0.0.1

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

INFO=$(GREEN)[INFO] $(RESET)
ERROR=$(RED)[ERROR] $(RESET)
WARN=$(YELLOW)[WARN] $(RESET)

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


#
# Go Dependencies commands
#

update-deps-featuretest:
	@echo "$(INFO) Updating dependencies for featuretest environment"
	cp featuretest.yaml glide.lock
	glide -y featuretest.yaml update
	cp glide.lock featuretest.lock

update-deps-master:
	@echo "$(INFO) Updating dependencies for $(BLUE)master$(RESET) environment"
	cp master.yaml glide.lock
	glide -y master.yaml update
	cp glide.lock master.lock

install-deps-featuretest:
	@echo "$(INFO) Installing dependencies for featuretest environment"
	cp featuretest.lock glide.lock
	glide -y featuretest.yaml install
	cp glide.lock featuretest.lock

install-deps-master:
	@echo "$(INFO) Installing dependencies for $(BLUE)master$(RESET) environment"
	cp master.lock glide.lock
	glide -y master.yaml install
	cp glide.lock master.lock


#
# Main
#

compile:
	@echo "$(INFO) Getting packages and building alpine go binary ..."
	@if [ "$(CURRENT_BRANCH)" != "master" && "$(CURRENT_BRANCH)" != "featuretest" ]; then \
		make update-deps-master; \
		make install-deps-master; \
	else \
		make update-deps-$(CURRENT_BRANCH); \
		make install-deps-$(CURRENT_BRANCH); \
	fi
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./app/cmd/addsvc/main.go

# TODO: Should speed this up with voluming vendor/
build-exec:        ##@build Cross compile the go binary executable
	@echo "$(INFO) Building a linux-alpine Go binary locally with a docker container $(BLUE)$(REPO):compile$(RESET)"
	docker build -t $(REPO):compile -f Dockerfile.build .
	docker run --rm -v "${PWD}":/go/src/github.com/newtonsystems/go-api-gateway $(REPO):compile
	@echo ""


#
# Run Commands
#
.PHONY: build run build-dev run-dev

build:                  ##@run Builds and run docker con
	docker build -t $(REPO):local .

run: build              ##@run Builds and run docker container with tag: '$(REPO):local' as a one-off.
	@echo "$(INFO) Running docker container with tag: $(REPO):local"
	@echo "$(BLUE)"
	@echo "$(INFO) Building docker container locally with tag: $(BLUE)$(REPO):local$(RESET)"

	docker run -it $(REPO):local
	@echo "$(NO_COLOR)"

build-dev:
	docker build -t $(REPO):dev -f Dockerfile.dev .

run-dev:    ##@dev Build and run (hot-reload) development docker container (Normally run this for dev changes)
	@echo ""
	docker run -v "${PWD}":/go/src/github.com/newtonsystems/go-api-gateway  -it $(REPO):dev




PID      = /tmp/awesome-golang-project.pid
GO_FILES = $(wildcard *.go)
APP      = ./main

serve: restart
	inotifywait -r -m . -e create -e modify | \
		while read path action file; do \
			echo "changed"; \
			make restart; \
		done

kill:
	kill `cat $(PID)` || true

before:
	echo "actually do nothing"

$(APP): $(GO_FILES)
	go build $? -o $@

restart: kill before $(APP)
	./main & echo $$! > $(PID)

.PHONY: serve restart kill before # let's go to reserve rules names







#
# Run Commands (Black Box)
#
.PHONY: run-latest-release run-latest

run-latest-release:     ##@run-black-box Run the current release (When you want to run as service as a black-box)
	@echo "$(INFO) Pulling release docker image for branch: newtonsystems/$(REPO):$(CURRENT_RELEASE_VERSION)"
	@echo "$(BLUE)"
	docker pull newtonsystems/$(REPO):$(CURRENT_RELEASE_VERSION);
	docker run newtonsystems/$(REPO):$(CURRENT_RELEASE_VERSION);
	@echo "$(NO_COLOR)"

run-latest:             ##@run-black-box Run the most up-to-date image for your branch from the docker registry or if the image doesnt exist yet you can specify. (When you want to run as service as a black-box)
	@echo "$(INFO) Running the most up-to-date image"
	@echo "$(INFO) Pulling latest docker image for branch: newtonsystems/$(REPO):$(CURRENT_BRANCH)"

	@docker pull newtonsystems/$(REPO):$(CURRENT_BRANCH); if [ $$? -ne 0 ] ; then \
		echo "$(ERROR) Failed to find image in registry: newtonsystems/$(REPO):$(CURRENT_BRANCH)"; \
		read -r -p "$(GREEN) Specific your own image name or Ctrl+C to exit:$(RESET)   " reply; \
		docker pull newtonsystems/$(REPO):$$reply; \
		docker run newtonsystems/$(REPO):$$reply; \
	else \
		docker run newtonsystems/$(REPO):$(CURRENT_BRANCH) app; \
	fi
