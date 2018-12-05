VERSION := ${shell git describe --always --long --dirty}
PKG := github.com/bhutch29/abv
DEPLOYPATH := /srv/http

all: install

install:
	go install -v -ldflags="-X main.version=${VERSION}" ./...

deploy:
	cp frontend/front.html ${DEPLOYPATH}/
	cp -r frontend/static ${DEPLOYPATH}/

run: install
	abv

.PHONY: run install deploy
