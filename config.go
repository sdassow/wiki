package main

const FileExtension = ".md"

type Git struct {
	url  string
	push bool
}

type Cookie struct {
	keyfile  string
	insecure bool
	key      []byte
}

type Listen struct {
	address  string
	network  string
	protocol string
}

type Tls struct {
	certfile string
	keyfile  string
}

// Config ...
type Config struct {
	listen   Listen
	prefix   string
	data     string
	brand    string
	bind     string
	git      Git
	cookie	 Cookie
	indexdir string
	tls      Tls
	hosts    []string
}
