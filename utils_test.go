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
		{"/WhatEver/", "/WhatEver/"},
		{"What/Ever", "What/Ever"},
		{"What/EverEver", "[What/EverEver](/view/What/EverEver)"},
		{"WhatEver, WhenEver", "[WhatEver](/view/WhatEver), [WhenEver](/view/WhenEver)"},
		{"WhatEver! WhenEver", "[WhatEver](/view/WhatEver)! [WhenEver](/view/WhenEver)"},
		{"WhatEver? WhenEver", "[WhatEver](/view/WhatEver)? [WhenEver](/view/WhenEver)"},
		{"WhatEver% WhenEver", "[WhatEver](/view/WhatEver)% [WhenEver](/view/WhenEver)"},
		{"WhatEver& WhenEver", "[WhatEver](/view/WhatEver)& [WhenEver](/view/WhenEver)"},
		{"WhatEver/WhenEver", "[WhatEver/WhenEver](/view/WhatEver/WhenEver)"},
		{"WhatEver/ WhenEver", "WhatEver/ [WhenEver](/view/WhenEver)"},
		{"WhatEver / WhenEver", "[WhatEver](/view/WhatEver) / [WhenEver](/view/WhenEver)"},
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
			t.Errorf("mismatch:\n   got:<%s>\n  !=\n  want:<%s>", out, check.out)
		}
	}

}

func TestCleanNewlines(t *testing.T) {
	checks := []struct {
		in  string
		out string
	}{
		{"foo\r\nbar\r\nbaz\r\n", "foo\nbar\nbaz\n"},
		{"foo\rbar\rbaz\r", "foo\nbar\nbaz\n"},
		{"foo\nbar\nbaz\n", "foo\nbar\nbaz\n"},
		{"foo\nbar\nbaz", "foo\nbar\nbaz\n"},
		{"foo\r\nbar\nbaz\r", "foo\nbar\nbaz\n"},
		{"foo\nbar\rbaz\n\n", "foo\nbar\nbaz\n\n"},
	}

	for _, check := range checks {
		out := CleanNewlines(check.in)
		if string(out) != check.out {
			t.Errorf("mismatch:\n  <%s>\n  !=\n  <%s>", out, check.out)
		}
	}

}
