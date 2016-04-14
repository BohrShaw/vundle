package main

import (
	"regexp"
	"testing"
)

func TestBundleDecode(t *testing.T) {
	bs := []string{"foo/bar", "foo/bar/baz", "a/b:", "a/b:br", "a/b:br/dir/dir"}
	f := regexp.MustCompile(`^[[:word:]-.]+/[[:word:]-.:]+$`)
	for _, b := range bs {
		if !f.MatchString(bundleDecode(b).repo) {
			t.Fatal("Bundle format %v is mis-decoded.", b)
		}
	}
}
