.PHONY: dev build release test clean

DATA_PATH = ../data

all: dev

dev: build
	./wiking --data ./checkout/ --git-url ./repo.git/

build: clean
	go build

test:
	go test ./...

clean:
	rm -rf wiking

fix-easymde-static-path:
	perl -pi.bak -e 's!(open\("GET",")[^"]*?(/en_US.(?:aff|dic)")!\1${DATA_PATH}\2!g' static/js/easymde.min.js
