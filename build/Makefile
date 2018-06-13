PROJECT := terraform-provider-signalform

ifdef BUILD_NUMBER
	REAL_BUILD_NUMBER?=$(BUILD_NUMBER)
else ifdef upstream_build_number
	REAL_BUILD_NUMBER=$(upstream_build_number)
else
	REAL_BUILD_NUMBER?=0
endif
ORG ?= default

VERSION := 3.0.0
TF_VERSION := 0.11
ITERATION := $(ORG)$(REAL_BUILD_NUMBER)

TF_PATH ?= /nail/opt/terraform-$(TF_VERSION)

PACKAGE_NAME := $(PROJECT)-$(TF_VERSION)_$(VERSION)-$(ITERATION)_amd64.deb
PACKAGE_FILE := dist/$(PACKAGE_NAME)

BINARY_FILE := dist/$(PROJECT)-$(GOOS)_$(GOARCH)

DOCKER_TAG = $(PROJECT)_$(shell date +%s)

CLEAN_CONTAINER := [ -e .docker_container_id ] && docker rm --force $$(cat .docker_container_id) || true; rm -f .docker_container_id


.PHONY: itest_%
itest_%: $(PACKAGE_FILE)
ifeq ($(ORG),yelp)
	$(eval DOCKER_BASE := docker-dev.yelpcorp.com/$*_yelp:latest)
else
	$(eval DOCKER_BASE := ubuntu:$*)
endif
	docker run --rm -v $(CURDIR)/../dist:/dist:ro -v $(CURDIR)/itest.sh:/itest.sh:ro $(DOCKER_BASE) bash /itest.sh /$(PACKAGE_FILE) $(TF_PATH)

$(PACKAGE_FILE): .docker_container_id
	test -d ../dist || mkdir ../dist/
	docker cp $$(cat .docker_container_id):/$(PACKAGE_FILE) ../dist/

.PHONY: binary
binary: $(BINARY_FILE)

# TODO: if GOOS and GOARCH change, we need to invalidate .docker_container_id
$(BINARY_FILE): .docker_container_id
	test -d ../dist || mkdir ../dist/
	docker cp $$(cat .docker_container_id):/$(BINARY_FILE) ../dist/

.docker_container_id: .docker_image_id
	docker run --rm=false \
		-e GOOS=$(GOOS) -e GOARCH=$(GOARCH) \
		-v $(CURDIR)/../src/terraform-provider-signalform:/go/src/terraform-provider-signalform:ro \
		-v $(CURDIR)/build.sh:/build.sh:ro \
		--cidfile=$(CURDIR)/.docker_container_id \
		$$(cat .docker_image_id) \
		bash /build.sh $(PROJECT) $(VERSION) $(ITERATION) $(TF_VERSION) $(TF_PATH) || \
	(retval=$$?; $(CLEAN_CONTAINER); exit $$retval; )
	docker logs $$(cat .docker_container_id)

.docker_image_id: Dockerfile
	docker build -t $(DOCKER_TAG) .
	docker inspect -f '{{ .Id }}' $(DOCKER_TAG) > .docker_image_id

.PHONY: clean
clean:
	$(CLEAN_CONTAINER)
	rm -f .docker_image_id
	rm -rf ../dist

.PHONY: shell
shell: .docker_image_id
	docker run --rm -t -i $$(cat .docker_image_id) /bin/bash

.PHONY: changelog
changelog:
	if [ ! -f changelog ]; then \
	    dch -v ${VERSION} --create -c changelog --package=$(PROJECT) -D trusty -u low ${ARGS}; \
	else \
	    dch -v ${VERSION} -c changelog -D trusty -u low ${ARGS}; \
	fi
	git add changelog
