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
		{"fobar WhatEver bla", "fobar [WhatEver](/view/WhatEver) bla"},
		{" - [OpenBSD Router : Native IPv6](https://lipidity.com/openbsd/router/) ",
			" - [OpenBSD Router : Native IPv6](https://lipidity.com/openbsd/router/) "},
		{" Another example [Quickstart code](/view/GoLang/QuickStart), doh ",
			" Another example [Quickstart code](/view/GoLang/QuickStart), doh "},
		{"Foo/WhatEver", "[Foo/WhatEver](/view/Foo/WhatEver)"},
		{"FooBar/WhatEver", "[FooBar/WhatEver](/view/FooBar/WhatEver)"},
		{"Foo/Bar/What/Ever", "Foo/Bar/What/Ever"},
	}

	for _, check := range checks {
		out := AutoCamelCase([]byte(check.in), "/view/")
		if string(out) != check.out {
			t.Errorf("mismatch:\n  <%s>\n  !=\n  <%s>", out, check.out)
		}
	}

}
