package main

import (
	"regexp"
	"testing"
)

func TestBundleDecode(t *testing.T) {
	bs := []string{"foo/bar", "foo/bar/baz", "a/b:", "a/b:br", "a/b:br/dir/dir"}
	fr := regexp.MustCompile(`^[[:word:]-.]+/[[:word:]-.]+$`)
	fb := regexp.MustCompile(`^[[:word:]-.]*$`)
	for _, b := range bs {
		bd := bundleDecode(b)
		if !fr.MatchString(bd.repo) || !fb.MatchString(bd.branch) {
			t.Errorf("Bundle format %v is mis-decoded as %v.", b, bd)
		}
	}
}
