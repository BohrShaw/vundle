package main

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestBundlesRaw(t *testing.T) {
	f, err := filepath.Abs("assets/bundles.txt")
	if err != nil {
		t.Error(err)
	}
	bundles := BundlesRaw(f)
	t.Log(bundles)
	count := len(bundles)
	if count != 12 {
		t.Errorf("The number of distinct bundles should be 12, but get %v.", count)
	}
	for _, b := range bundles {
		if !regexp.MustCompile(`^[[:word:]-.]+/[[:word:]-.:/]+$`).MatchString(b) {
			t.Errorf("Raw bundle %v is of wrong format.", b)
		}
	}
}

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
