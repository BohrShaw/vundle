package main

import (
	"regexp"
	"testing"
)

func TestBundleDecode(t *testing.T) {
	bs := []string{"foo/bar", "foo/bar/baz", "foo/bar/baz/bas",
		"a/b:", "a/b:br", "a/b:br/dir/dir", "a/b:br/dir/dir/"}
	fr := regexp.MustCompile(`^[[:word:]-.]+/[[:word:]-.]+$`)
	fb := regexp.MustCompile(`^[[:word:]-.]*$`)
	for _, b := range bs {
		bd, err := bundleDecode(b)
		if err != nil {
			t.Errorf("Bundle format %v is unrecognized.", b)
			continue
		}
		if !fr.MatchString(bd.repo) || !fb.MatchString(bd.branch) {
			t.Errorf("Bundle format %v is mis-decoded as %v.", b, bd)
		}
	}
}
