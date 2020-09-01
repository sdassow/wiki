.PHONY: dev build release test clean

all: dev

dev: build
	./wiking --data ./checkout/ --git-url ./repo.git/

build: clean
	go build

test:
	go test ./...

clean:
	rm -rf wiking
