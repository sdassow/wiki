# wiki
[![GoDoc](https://godoc.org/github.com/openmicroapps/wiki?status.svg)](https://godoc.org/github.com/openmicroapps/wiki)
[![Go Report Card](https://goreportcard.com/badge/github.com/openmicroapps/wiki)](https://goreportcard.com/report/github.com/openmicroapps/wiki)

wiki is a self-hosted well uh wiki engine or content management system that
lets you create and share content in Markdown format.

### Source

```#!bash
$ go get github.com/openmicroapps/wiki
```

## Usage

Run wiki:

```#!bash
$ wiki
```

Visit: http://localhost:8000/

Start creating/editing content!

## Configuration

By default wiki pages are stored in `./data` in the local directory. This can
be changed by supplying the `-data /path/to/data` option.

## License

MIT
