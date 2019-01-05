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

dockerbuild:
	docker build -t abv_api:latest -f api/Dockerfile .
	docker build -t abv_frontend:latest -f api/Dockerfile .

dockerrun:
	docker run -d -p 8081:8081 --name abv_api abv_api
	docker run -d -p 80:8080 --name abv_frontend abv_frontend

.PHONY: run install deploy
