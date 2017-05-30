PACKAGE := terraform-provider-signalform
GOPATH  := $(shell pwd -L)
BASE    := $(GOPATH)/src/$(PACKAGE)
PATH    := $(GOPATH)/bin:$(PATH)
GLIDE   := glide
export GOPATH
export PATH
# unset GOROOT avoids: "go test error: cannot use matchString as type testing.testDeps in argument to testing.MainStart"
unexport GOROOT

.PHONY: all fmt terraform-provider-ddns clean package test itest_%
all: fmt .git/hooks/pre-commit test build

fmt:
	go fmt ./...

deps:
	@echo Getting dependencies...
	@go get github.com/Masterminds/glide
	@cd src/github.com/Masterminds/glide && git checkout --quiet v0.12.3
	@go build -o bin/glide github.com/Masterminds/glide/
	@cd $(BASE) && $(GLIDE) install

clean:
	rm -rf bin
	rm -rf pkg
	make -C yelppack clean

build: test
	mkdir -p $(GOPATH)/bin
	cd $(BASE) && go build -o $(GOPATH)/bin/terraform-provider-signalform
	cp /nail/opt/terraform-0.8/bin/terraform bin/

integration:
	make -C test

itest_%:
	mkdir -p dist
	make -C yelppack $@

package: itest_lucid

test: deps
	cd $(BASE) && go test -v $$(glide novendor)

itest_%:
	make -C yelppack $@

.git/hooks/pre-commit:
	if [ ! -f .git/hooks/pre-commit ]; then ln -s ../../git-hooks/pre-commit .git/hooks/pre-commit; fi
