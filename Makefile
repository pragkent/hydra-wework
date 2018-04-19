GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_DIRTY := $(shell test -n "`git status --porcelain`" && echo "*" || true)
BUILD_VERSION := $(shell cat VERSION)
BUILD_TIME := $(shell date +%FT%T%z)

LDFLAGS := " \
-X main.buildGitCommit=${GIT_COMMIT}${GIT_DIRTY} \
-X main.buildGitBranch=${GIT_BRANCH} \
-X main.buildVersion=${BUILD_VERSION} \
-X main.buildTime=${BUILD_TIME} \
"

bin:
	go build -ldflags ${LDFLAGS} -o bin/hydra-wework

docker-bin:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o bin/hydra-wework

fmt:
	go fmt ./...

test:
	go test ./...

testcover:
	go test -cover ./...

testrace:
	go test -race ./...

vet:
	go vet ./...

dep:
	dep -ensure -v

clean:
	rm -f bin/

.PHONY: bin docker-bin fmt test cover testrace vet clean
