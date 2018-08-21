// Copyright 2018 Schibsted

package buildbuild

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

type Scanner struct {
	*bufio.Scanner
	io.Closer
	Filename string

	scannerSpecials map[rune]bool
	scannerNewword  bool
}

var (
	builddescSpecials = map[rune]bool{'(': true, ')': true, '[': true, ']': true, ':': true}
	argsSpecials      = map[rune]bool{'[': true, ']': true}
)

func NewScanner(r io.ReadCloser, file string) *Scanner {
	s := &Scanner{Scanner: bufio.NewScanner(r), Closer: r, Filename: file}
	s.Split(s.Splitter)
	s.scannerSpecials = builddescSpecials
	return s
}

func (s *Scanner) Splitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces, return special tokens, otherwise scan for words.
	start := 0
	var r rune
	var width int
	for {
		if start >= len(data) {
			return start, nil, nil
		}
		r, width = utf8.DecodeRune(data[start:])
		// Check for comments, runs from # to end of line.
		if r == '#' {
			for start < len(data) {
				r, width = utf8.DecodeRune(data[start:])
				if r == '\n' {
					break
				}
				start += width
			}
		} else if !unicode.IsSpace(r) {
			break
		}
		start += width
		s.scannerNewword = true
	}
	if s.scannerSpecials[r] {
		return start + width, data[start : start+width], nil
	}
	end := start + width
	for end < len(data) {
		r, width = utf8.DecodeRune(data[end:])
		if unicode.IsSpace(r) || s.scannerSpecials[r] {
			return end, data[start:end], nil
		}
		end += width
	}
	if atEOF {
		return len(data), data[start:], nil
	}
	return start, nil, nil
}
