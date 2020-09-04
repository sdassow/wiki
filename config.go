package main

const FileExtension = ".md"

type Git struct {
	url  string
	push bool
}

type Csrf struct {
	keyfile  string
	insecure bool
	key      []byte
}

// Config ...
type Config struct {
	data  string
	brand string
	bind  string
	git   Git
	csrf  Csrf
}
