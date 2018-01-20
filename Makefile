# If nothing was specified, run all targets as if in a fresh clone
.PHONY: all
## Default target
all: start-db build kill-db clean

POSTGRES_PASSWORD:=mysecretpassword
UNAME_S := $(shell uname -s)


.PHONY: init
init:
	@rm -rf .tmp
	@mkdir .tmp

.PHONY: clean
clean:
	@rm -rf .tmp

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
	docker build --build-arg POSTGRES_HOST=`cat .tmp/postgres.host` \
	  --build-arg POSTGRES_PORT=`cat .tmp/postgres.port` \
	  --build-arg POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	  . \
	  -t xcoulon/go-url-shortener:latest

.PHONY: kill-db
kill-db:
	@echo "killing the test db container..."
	docker rm -f `cat .tmp/postgres.cid`