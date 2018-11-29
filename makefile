VERSION := ${shell git describe --always --long --dirty}
PKG := github.com/bhutch29/abv

all: install

install:
	go install -v -ldflags="-X main.version=${VERSION}" ./...

run: install
	abv

.PHONY: run install
