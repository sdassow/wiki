package main

import (
	"testing"
)

func TestAutoCamelCase(t *testing.T) {
	checks := []struct {
		in  string
		out string
	}{
		{"WhatEver", "[WhatEver](/view/WhatEver)"},
		{"[AnExampleLink](http://example.com)", "[AnExampleLink](http://example.com)"},
	}

	for _, check := range checks {
		out := AutoCamelCase([]byte(check.in), "/view/")
		if string(out) != check.out {
			t.Errorf("mismatch:\n  <%s>\n  !=\n  <%s>", out, check.out)
		}
	}

}
