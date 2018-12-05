VERSION := ${shell git describe --always --long --dirty}
PKG := github.com/bhutch29/abv

all: install

install:
	go install -v -ldflags="-X main.version=${VERSION}" ./...

deploy:
	cp frontend/front.html /srv/http/
	cp -r frontend/static /srv/http/

run: install
	abv

.PHONY: run install deploy
