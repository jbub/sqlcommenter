package sqlcommenter

// Code is adapted from standard library package net/url.
// Copyright (c) 2009 The Go Authors. All rights reserved.

import (
	"bytes"
	"testing"
)

func TestQueryEscape(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			in:   "",
			want: "",
		},
		{
			in:   "abc",
			want: "abc",
		},
		{
			in:   "one two",
			want: "one+two",
		},
		{
			in:   "10%",
			want: "10%25",
		},
		{
			in:   " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			want: "+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B",
		},
	}

	for _, cs := range cases {
		t.Run(cs.in, func(t *testing.T) {
			var b bytes.Buffer
			writeQueryEscape(cs.in, &b)
			if got := b.String(); cs.want != got {
				t.Errorf("got %q, want %q", got, cs.want)
			}
		})
	}
}

func TestPathEscape(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			in:   "",
			want: "",
		},
		{
			in:   "abc",
			want: "abc",
		},
		{
			in:   "abc+def",
			want: "abc+def",
		},
		{
			in:   "a/b",
			want: "a%2Fb",
		},
		{
			in:   "one two",
			want: "one%20two",
		},
		{
			in:   "10%",
			want: "10%25",
		},
		{
			in:   " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			want: "%20%3F&=%23+%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09:%2F@$%27%28%29%2A%2C%3B",
		},
	}

	for _, cs := range cases {
		t.Run(cs.in, func(t *testing.T) {
			var b bytes.Buffer
			writePathEscape(cs.in, &b)
			if got := b.String(); cs.want != got {
				t.Errorf("got %q, want %q", got, cs.want)
			}
		})
	}
}
