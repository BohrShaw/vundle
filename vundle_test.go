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
				protSSH := strings.HasSuffix(d, ":")
				if protSSH && r.prefix == "git@" || !protSSH && r.prefix == "https://" {
				} else {
					t.Errorf("'%v' is mis-decoded with prefix '%v'", c, r.prefix)
				}
				if d == "" && r.domain == "github.com/" || d != "" && r.domain == d {
				} else {
					t.Errorf("'%v' is mis-decoded with domain '%v'", c, r.domain)
				}
				if r.repo != repo {
					t.Errorf("'%v' is mis-decoded with repo '%v'", c, r.repo)
				}
				if b == ":branch" && r.branch == "branch" || b == ":" && r.branch == PLATFORM || b == "" && r.branch == "" {
				} else {
					t.Errorf("'%v' is mis-decoded with branch '%v'", c, r.branch)
				}
			}
		}
	}
}
