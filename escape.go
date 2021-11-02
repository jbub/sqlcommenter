package sqlcommenter

// Code is adapted from standard library package net/url.
// Copyright (c) 2009 The Go Authors. All rights reserved.

import (
	"bytes"
)

const upperhex = "0123456789ABCDEF"

func writeQueryEscape(s string, b *bytes.Buffer) {
	writeEscape(s, true, b)
}

func writePathEscape(s string, b *bytes.Buffer) {
	writeEscape(s, false, b)
}

func writeEscape(s string, query bool, b *bytes.Buffer) {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, query) {
			if c == ' ' && query {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		b.WriteString(s)
		return
	}

	if hexCount == 0 {
		for i := 0; i < len(s); i++ {
			if s[i] == ' ' {
				b.WriteByte('+')
			} else {
				b.WriteByte(s[i])
			}
		}
		return
	}

	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && query:
			b.WriteByte('+')
		case shouldEscape(c, query):
			b.WriteByte('%')
			b.WriteByte(upperhex[c>>4])
			b.WriteByte(upperhex[c&15])
		default:
			b.WriteByte(c)
		}
	}
}

func shouldEscape(c byte, query bool) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '_', '.', '~':
		return false
	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@':
		if query {
			return true
		}
		return c == '/' || c == ';' || c == ',' || c == '?'
	}
	return true
}
