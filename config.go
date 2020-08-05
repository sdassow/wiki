package main

const FileExtension = ".md"

type Git struct {
	url  string
	push bool
}

// Config ...
type Config struct {
	data  string
	brand string
	bind  string
	git   Git
}
