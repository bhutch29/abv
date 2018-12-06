VERSION := ${shell git describe --always --long --dirty}
PKG := github.com/bhutch29/abv
DEPLOYPATH := /srv/http

all: test-quiet install

test-quiet:
	go test ./...

test:
	go test -v ./...

install:
	go install -v -ldflags="-X main.version=${VERSION}" ./...

deploy:
	cp frontend/front.html ${DEPLOYPATH}/
	cp -r frontend/static ${DEPLOYPATH}/

run: install
	abv

.PHONY: run install deploy
