SHELL = bash

GOTAGS ?=

GOFILES ?= $(shell go list ./... | grep -v /vendor/)

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)

all: dev
	echo ${GOFILES}

bin:
	@mkdir -p bin/
	@GOTAGS='$(GOTAGS)' sh -c "'$(CURDIR)/scripts/build.sh'"

dev:
	@echo "--> Building stor"
	mkdir -p pkg/$(GOOS)_$(GOARCH)/ bin/
	go install -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)'
	cp $(GOPATH)/bin/stor bin/
	cp $(GOPATH)/bin/stor pkg/$(GOOS)_$(GOARCH)

# linux builds a linux package independent of the source platform
linux:
	mkdir -p pkg/linux_amd64/
	GOOS=linux GOARCH=amd64 go build -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/linux_amd64/stor

format:
	@echo "--> Running go fmt"
	@go fmt $(GOFILES)

vet:
	@echo "--> Running go vet"
	@go vet -tags '$(GOTAGS)' $(GOFILES); if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

.PHONY: all bin dev dist format vet
