OUT := abv
VERSION := ${shell git describe --always --long --dirty}
PKG := github.com/bhutch29/abv

all: run

build:
	go build -i -v -o ${OUT} -ldflags="-X main.version=${VERSION}" ${PKG}

run: build
	./${OUT}

.PHONY: run build
