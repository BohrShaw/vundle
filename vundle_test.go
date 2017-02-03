package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBundlesRaw(t *testing.T) {
	f, err := filepath.Abs("assets/bundles.txt")
	if err != nil {
		t.Error(err)
	}
	bundles := BundlesRaw([]string{f})
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
	repo := "author/project"
	for _, d := range []string{"domain.com/", "sub.domain.com/", "domain.com:", "sub.domain.com:", ""} {
		for _, b := range []string{":branch", ":", ""} {
			for _, s := range []string{"/sub/directory", "/sub-directory", ""} {
				c := d + repo + b + s
				r, err := bundleDecode(c)
				if err != nil {
					t.Errorf("Bundle '%v' is unrecognized while it should be.", r)
					continue
				}
				var msg string
				protSSH := strings.HasSuffix(d, ":")
				if protSSH && r.prefix == "git@" || !protSSH && r.prefix == "https://" {
				} else {
					msg += " prefix \"" + r.prefix + "\""
				}
				if d == "" && r.domain == "github.com/" || d != "" && r.domain == d {
				} else {
					msg += " domain \"" + r.domain + "\""
				}
				if r.repo != repo {
					msg += " repo \"" + r.repo + "\""
				}
				if b == ":branch" && r.branch == "branch" || b == ":" && r.branch == PLATFORM || b == "" && r.branch == "" {
				} else {
					msg += " branch \"" + r.branch + "\""
				}
				if msg != "" {
					t.Errorf("Bundle \"%v\" is mis-decoded with" + msg + ".", c)
				}
			}
		}
	}
}
