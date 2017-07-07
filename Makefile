.PHONY: dev build clean

all: dev

dev: build
	./wiki

build: clean
	go get ./...
	go build .

test:
	go test ./...

clean:
	rm -rf wiki
