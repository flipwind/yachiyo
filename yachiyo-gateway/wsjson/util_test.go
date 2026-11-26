package wsjson

import (
	"strings"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	valid := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"group/168", "group/168"},
		{"/group/168/", "group/168"},
		{"qq.group_1-a/room_2", "qq.group_1-a/room_2"},
		{strings.Repeat("a", maxPathLength), strings.Repeat("a", maxPathLength)},
	}

	for _, tc := range valid {
		got, err := normalizePath(tc.in)
		if err != nil {
			t.Fatalf("normalizePath(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("normalizePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	invalid := []string{
		"/",
		"///",
		"group//123",
		"group/../123",
		"..",
		".",
		"group/./123",
		"group/123?foo=bar",
		"group/123#frag",
		"group/hello world",
		"group/中文",
		strings.Repeat("a", maxPathLength+1),
	}

	for _, in := range invalid {
		if got, err := normalizePath(in); err == nil {
			t.Fatalf("normalizePath(%q) = %q, want error", in, got)
		}
	}
}
