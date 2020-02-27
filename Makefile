PKG_NAME ?= github.com/chenchun/vms
VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X $(PKG_NAME)/pkg/version.gitVersion=$(VERSION) -X $(PKG_NAME)/pkg/version.gitCommit=$(GIT_COMMIT) -X $(PKG_NAME)/pkg/version.buildDate=$(BUILD_DATE)
GOBIN ?= $(GOPATH)/bin
IMAGE := chenchun/vms:$(VERSION)
DOCKER_BUILD := docker run --rm -v $(GOPATH)/pkg:/go/pkg -v $(shell pwd):/go/src/$(PKG_NAME) --workdir=/go/src/$(PKG_NAME) golang:1.13

.PHONY: verify
verify:
	$(eval unformatted = $(shell gofmt -s -l main.go pkg/))
	@$(if $(strip $(unformatted)),(echo "please gofmt $(unformatted)"; exit 1),)

.PHONY: build
build: verify
	CGO_ENABLED=0 go build -o ./bin/vms -ldflags "$(LDFLAGS)" .

.PHONY: docker-build
docker-build:
ifeq ("$(shell go env GOOS)", "darwin")
	$(DOCKER_BUILD) make build
else
	make build
endif

.PHONY: docker
docker: docker-build
	docker build -f Dockerfile -t "$(IMAGE)" .

.PHONY: push
push: docker
	docker push "$(IMAGE)"

.PHONY: clean
clean:
	rm -f bin/*