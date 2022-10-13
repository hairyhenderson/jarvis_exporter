.DEFAULT_GOAL = build
BIN_DIR := bin
PKG_NAME := jarvis_exporter
DOCKER_REPO ?= hairyhenderson/$(PKG_NAME)
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64,linux/arm/v6,linux/arm/v7
# we just build by default, as a "dry run"
BUILDX_ACTION ?= --output=type=image,push=false
# BUILDX_ACTION ?= --push
TAG_LATEST ?= latest

GOOS ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/')
GOARCH ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/')

ifeq ("$(TARGETVARIANT)","")
ifneq ("$(GOARM)","")
TARGETVARIANT := v$(GOARM)
endif
else
ifeq ("$(GOARM)","")
GOARM ?= $(subst v,,$(TARGETVARIANT))
endif
endif

build: bin/jarvis_exporter

bin/%: $(shell find . -type f -name "*.go") go.mod go.sum
	go build \
		-ldflags "-w -s" \
		-o $@ \
		./cmd/$(patsubst bin/%,%,$@)

docker-multi: Dockerfile
	docker buildx build \
		--build-arg VCS_REF=$(COMMIT) \
		--build-arg PKG_NAME=$(PKG_NAME) \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_REPO):$(TAG_LATEST) \
		--target runtime \
		$(BUILDX_ACTION) .

test:
	@go test -race -coverprofile=c.out ./...

lint:
	@golangci-lint run -v --max-same-issues=0 --max-issues-per-linter=0

ci-lint:
	@golangci-lint run -v --max-same-issues=0 --max-issues-per-linter=0 --out-format=github-actions

.PHONY: test lint ci-lint
.DELETE_ON_ERROR:
.SECONDARY:
