APP_NAME = terraform-provider-passbolt
VERSION = 1.0.0

.PHONY: help
help:
	@echo "make options\n\
		- all                         deps, vet, fmt, build, docs, install\n\
		- deps                        fetch all dependencies\n\
		- build                       build binary ${APP_NAME}\n\
		- docs                        generate tf docs\n\
		- install                     install binary ${APP_NAME} in gopath\n\
		- help                        display this message"

.PHONY: all
all: deps vet fmt build docs install

.PHONY: deps
deps:
	go mod tidy

.PHONY: docs
docs:
	./scripts/generate_docs.sh

.PHONY: build
build:
	go build

.PHONY: install
install:
	go install

.PHONY: test
test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

