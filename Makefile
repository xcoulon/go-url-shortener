# If nothing was specified, show help
.PHONY: help
# Based on https://gist.github.com/rcmachado/af3db315e31383502660
## Display this help text.
help:/
	$(info Available targets)
	$(info -----------------)
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		helpCommand = substr($$1, 0, index($$1, ":")-1); \
		if (helpMessage) { \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub(/##/, "\n                                     ", helpMessage); \
			printf "%-35s - %s\n", helpCommand, helpMessage; \
			lastLine = "" \
		} \
	} \
	{ hasComment = match(lastLine, /^## (.*)/); \
          if(hasComment) { \
            lastLine=lastLine$$0; \
	  } \
          else { \
	    lastLine = $$0 \
          } \
        }' $(MAKEFILE_LIST)

POSTGRES_PASSWORD:=mysecretpassword
UNAME_S := $(shell uname -s)
OC_USERNAME := developer
OC_PASSWORD := developer
MINISHIFT_IP := $(shell minishift ip)
PROJECT_NAME := sandbox
OC_PROJECT=$(PROJECT_NAME)
REGISTRY_URI := $(shell minishift openshift registry)
REGISTRY_IMAGE = webapp

.PHONY: init
init:
	@rm -rf .tmp
	@mkdir .tmp


.PHONY: clean
clean:
	@rm -rf .tmp

.PHONY: minishift-login
## login to minishift
minishift-login:
	@echo "Login to minishift..."
	@oc login --insecure-skip-tls-verify=true https://$(MINISHIFT_IP):8443  -u developer -p developer 1>/dev/null

.PHONY: minishift-deploy
## deploy the image on minishift
minishift-deploy: minishift-login 
	eval $$(minishift docker-env) && docker login -u developer -p $(shell oc whoami -t) $(shell minishift openshift registry) && docker tag ${OC_PROJECT}/${REGISTRY_IMAGE}:latest ${REGISTRY_URI}/${OC_PROJECT}/${REGISTRY_IMAGE}:latest
	eval $$(minishift docker-env) && docker login -u developer -p $(shell oc whoami -t) $(shell minishift openshift registry) && docker push $(shell minishift openshift registry)/${OC_PROJECT}/${REGISTRY_IMAGE}:latest

.PHONY: minishift-image
## build image on minishift
minishift-image:
	echo "building the application image..."
	$(eval BUILD_COMMIT:=$(shell git rev-parse --short HEAD))
	$(eval BUILD_TIME:=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	eval $$(minishift docker-env) && \
	docker build --build-arg POSTGRES_HOST=`cat .tmp/postgres.host` \
	--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
	--build-arg BUILD_TIME=$(BUILD_TIME) \
	--file Dockerfile.openshift \
	. \
	-t ${OC_PROJECT}/$(REGISTRY_IMAGE):$(BUILD_COMMIT) && \
	docker tag ${OC_PROJECT}/$(REGISTRY_IMAGE):$(BUILD_COMMIT) ${OC_PROJECT}/$(REGISTRY_IMAGE):latest

.PHONY: start-db
start-db: init
	@echo "starting the test db container..."
ifeq ($(UNAME_S),Darwin)
	@echo "docker.for.mac.host.internal" > .tmp/postgres.host
else
	@echo "localhost" > .tmp/postgres.host
endif
	docker run -P -d --cidfile .tmp/postgres.cid -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) postgres:10.1 > /dev/null
	docker inspect `cat .tmp/postgres.cid` \
	  --format='{{ with index .NetworkSettings.Ports "5432/tcp" }}{{ with index . 0 }}{{ index . "HostPort" }}{{ end }}{{ end }}' \
	  > .tmp/postgres.port
	@echo "test db instance is listening on `cat .tmp/postgres.host`:`cat .tmp/postgres.port`"

.PHONY: build
build: start-db
	@echo "building the application image..."
	$(eval BUILD_COMMIT:=$(shell git rev-parse --short HEAD))
	$(eval BUILD_TIME:=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	docker build --build-arg POSTGRES_HOST=`cat .tmp/postgres.host` \
	  --build-arg POSTGRES_PORT=`cat .tmp/postgres.port` \
	  --build-arg POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	  --build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  --file Dockerfile.openshift \
	  . \
	  -t $(REGISTRY_IMAGE):$(BUILD_COMMIT)
	docker tag $(REGISTRY_IMAGE):$(BUILD_COMMIT) $(REGISTRY_IMAGE):latest

.PHONY: kill-db
kill-db:
	@echo "killing the test db container..."
	docker rm -f `cat .tmp/postgres.cid`