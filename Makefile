.PHONY: dev build release test clean

DATA_PATH = ../data

all: dev

dev: build
	./wiki

build: clean
	go get ./...
	go build .

release:
	@./tools/release.sh

test:
	go test ./...

clean:
	rm -rf wiki

fix-easymde-static-path:
	perl -pi.bak -e 's!(open\("GET",")[^"]*?(/en_US.(?:aff|dic)")!\1${DATA_PATH}\2!g' static/js/easymde.min.js
