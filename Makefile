PKG = github.com/nathan-osman/influx-daylight
CMD = influxdl

CWD = $(shell pwd)
UID = $(shell id -u)
GID = $(shell id -g)

# Find all Go source files (excluding the cache path)
SOURCES = $(shell find -type f -name '*.go' ! -path './cache/*')

all: dist/${CMD}

# Build the standalone executable
dist/${CMD}: ${SOURCES} | cache/lib cache/src/${PKG} dist
	@docker run \
	    --rm \
	    -e CGO_ENABLED=0 \
	    -e GIT_COMMITTER_NAME=a \
	    -e GIT_COMMITTER_EMAIL=b \
	    -u ${UID}:${GID} \
	    -v ${CWD}/cache/lib:/go/lib \
	    -v ${CWD}/cache/src:/go/src \
	    -v ${CWD}/dist:/go/bin \
	    -v ${CWD}:/go/src/${PKG} \
	    golang:latest \
	    go get -pkgdir /go/lib ${PKG}/cmd/${CMD}

cache/lib:
	@mkdir -p cache/lib

cache/src/${PKG}:
	@mkdir -p cache/src/${PKG}

dist:
	@mkdir dist

clean:
	@rm -rf cache dist

.PHONY: clean
