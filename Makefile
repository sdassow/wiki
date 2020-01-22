.PHONY: dev build release test clean

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
