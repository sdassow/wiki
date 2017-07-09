# wiki
[![GoDoc](https://godoc.org/github.com/prologic/wiki?status.svg)](https://godoc.org/github.com/prologic/wiki)
[![Go Report Card](https://goreportcard.com/badge/github.com/prologic/wiki)](https://goreportcard.com/report/github.com/prologic/wiki)

wiki is a self-hosted well uh wiki engine or content management system that
lets you create and share content in Markdown format.

### Source

```#!bash
$ go install github.com/prologic/wiki/...
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
