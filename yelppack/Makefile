.PHONY: itest_% clean shell
PROJECT = terraform-provider-signalform

VERSION = 2.2.3
ITERATION = yelp1
TF_VERSION = 0.9

PACKAGE_NAME := $(PROJECT)-$(TF_VERSION)_$(VERSION)-$(ITERATION)_amd64.deb
PACKAGE_FILE := dist/$(PACKAGE_NAME)

DOCKER_TAG = $(PROJECT)_$(shell date +%s)

CLEAN_CONTAINER := [ -e .docker_container_id ] && docker rm --force $$(cat .docker_container_id) || true; rm -f .docker_container_id

itest_%: $(PACKAGE_FILE)
	docker run --rm -v $(CURDIR)/../dist:/dist:ro -v $(CURDIR)/itest.sh:/itest.sh:ro docker-dev.yelpcorp.com/$*_yelp:latest bash /itest.sh /$(PACKAGE_FILE) $(TF_VERSION)

$(PACKAGE_FILE): .docker_container_id
	test -d ../dist || mkdir ../dist/
	docker cp $$(cat .docker_container_id):/$(PACKAGE_FILE) ../dist/

.docker_container_id: .docker_image_id
	docker run --rm=false \
		-v $(CURDIR)/../src/terraform-provider-signalform:/go/src/terraform-provider-signalform:ro \
		-v $(CURDIR)/build.sh:/build.sh:ro \
		--cidfile=$(CURDIR)/.docker_container_id \
		$$(cat .docker_image_id) \
		bash /build.sh $(PROJECT) $(VERSION) $(ITERATION) $(TF_VERSION) || \
	(retval=$$?; $(CLEAN_CONTAINER); exit $$retval; )
	docker logs $$(cat .docker_container_id)

.docker_image_id: Dockerfile
	docker build -t $(DOCKER_TAG) .
	docker inspect -f '{{ .Id }}' $(DOCKER_TAG) > .docker_image_id

clean:
	$(CLEAN_CONTAINER)
	rm .docker_image_id
	rm -rf ../dist

shell: .docker_image_id
	docker run --rm -t -i $$(cat .docker_image_id) /bin/bash
