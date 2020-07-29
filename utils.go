package main

import (
	"fmt"
	"regexp"
)

var validPageRegex = regexp.MustCompile(`(^|\s)((?:[A-Z][A-Za-z0-9]+/)*[A-Z][a-z]+[A-Z][A-Za-z0-9]+)(\s|$)`)

func AutoCamelCase(body []byte, base string) []byte {
	return validPageRegex.ReplaceAll(
		body,
		[]byte(fmt.Sprintf("$1[$2](%s$2)$3", base)),
	)
}
