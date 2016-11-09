.PHONY: all fmt .git/hooks/pre-commit terraform-provider-ddns clean package test itest_%

all: fmt .git/hooks/pre-commit test terraform-provider-signalform

fmt:
	go fmt ./...

clean:
	make -C yelppack clean
	rm -f terraform-provider-signalform
	rm -rf test/example.git test/checkout test/terraform.tfstate.backup test/terraform.tfstate

terraform-provider-signalform: test
	go build

dev: terraform-provider-signalform
	cp terraform-provider-signalform $$(echo $$GOPATH|sed -e's/://')/bin

integration: dev
	make -C test

itest_%:
	mkdir -p dist
	make -C yelppack $@

package: itest_lucid

test:
	go test -v ./signalform/...

itest_%:
		make -C yelppack $@


.git/hooks/pre-commit:
	if [ ! -f .git/hooks/pre-commit ]; then ln -s ../../git-hooks/pre-commit .git/hooks/pre-commit; fi

