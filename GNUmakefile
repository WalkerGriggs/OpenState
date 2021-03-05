SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
OS := $(shell uname | cut -d- -f1)
PREFIX := "  --"

ifeq (Linux,$(OS))
ALL_TARGETS += linux_amd64 \
               linux_386
endif

ifeq (Darwin,$(OS))
ALL_TARGETS += darwin_amd64
endif

##
## Darwin
##

pkg/darwin_amd64/openstate: $(SOURCE_FILES)
	@echo "$(PREFIX) Building $@"
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build \
		-trimpath \
		-o "$@"

##
## Linux
##

pkg/linux_amd64/openstate: $(SOURCE_FILES)
	@echo "$(PREFIX) Building $@"
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build \
		-trimpath \
		-o "$@"

pkg/linux_386/openstate: $(SOURCE_FILES)
	@echo "$(PREFIX) Building $@"
	@CGO_ENABLED=1 GOOS=linux GOARCH=386 \
		go build \
		-trimpath \
		-o "$@"

.PHONY: dev
dev: GOOS=$(shell go env GOOS)
dev: GOARCH=$(shell go env GOARCH)
dev: GOPATH=$(shell go env GOPATH)
dev: DEV_TARGET=pkg/$(GOOS)_$(GOARCH)/openstate
dev:
	@echo "$(PREFIX) Removing old development build"
	@rm -f $(PROJECT_ROOT)/$(DEV_TARGET)
	@rm -f $(PROJECT_ROOT)/bin/openstate
	@rm -f $(GOPATH)/bin/openstate
	@$(MAKE) --no-print-directory $(DEV_TARGET)
	@mkdir -p $(PROJECT_ROOT)/bin
	@mkdir -p $(GOPATH)/bin
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(PROJECT_ROOT)/bin/
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(GOPATH)/bin

.PHONY: all
all: clean $(foreach t,$(ALL_TARGETS),pkg/$(t)/openstate)
	@echo "$(PREFIX) Results:"
	@tree $(PROJECT_ROOT)/pkg

.PHONY: clean
clean: GOPATH=$(shell go env GOPATH)
clean:
	@echo "$(PREFIX) Cleaning builds"
	@rm -rf "$(PROJECT_ROOT)/bin/"
	@rm -rf "$(PROJECT_ROOT)/pkg/"
	@rm -f "$(GOPATH)/bin/openstate"

.PHONY: tidy
tidy:
	@echo "$(PREFIX) Tidying module"
	@go mod tidy

.PHONY: vendor
vendor: tidy
	@echo "$(PREFIX) Vendoring deps"
	@go mod vendor
