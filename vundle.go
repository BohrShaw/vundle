// Author: Bohr Shaw <pubohr@gmail.com>

// Vundle manages Vim bundles(plugins).
// It downloads, updates bundles, clean disabled bundles.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// Bundle specify the repository of format "author/project" and the branch.
type Bundle struct {
	repo, branch string
}

var (
	update           = flag.Bool("u", false, "update bundles")
	_filter          = flag.String("f", ".", "filter bundles with a go regexp")
	filter           = regexp.Regexp{}
	clean            = flag.Bool("c", false, "clean bundles")
	dryrun           = flag.Bool("n", false, "dry run (noop)")
	routines         = flag.Int("r", 12, "max number of routines")
	bundles          = Bundles()
	_user, _         = user.Current()
	home             = _user.HomeDir
	root             = home + "/.vim/bundle"
	git, gitNotExist = exec.LookPath("git")
	sep              = "============ "
)

func init() {
	flag.Parse()
	filter = *regexp.MustCompile(*_filter)
	if gitNotExist != nil {
		log.Fatal(gitNotExist)
	}
}

func main() {
	ch := make(chan Bundle, 9)
	wg := sync.WaitGroup{} // goroutines count
	routineCount := 0
	for _, b := range bundles {
		if filter.MatchString(b.repo) {
			ch <- b
			if routineCount <= *routines {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for b := range ch {
						Sync(b)
					}
				}()
				routineCount++
			}
		}
	}
	close(ch)

	wg.Wait()
	Helptags()

	if *clean {
		Clean()
	}
}

// Sync clone or update a bundle
func Sync(b Bundle) {
	cmd := &exec.Cmd{Path: git}
	url := "https://github.com/" + b.repo
	dest := root + "/" + strings.Split(b.repo, "/")[1]
	var output bytes.Buffer

	_, err := os.Stat(dest)
	if os.IsNotExist(err) { // clone
		args := make([]string, 0, 10)
		args = append(args, git, "clone", "--depth", "1", "--recursive", "--quiet")
		if b.branch != "" {
			args = append(args, "--branch", b.branch)
		}
		cmd.Args = append(args, url, dest)

		output.WriteString(sep + url + " ")
		err := cmd.Run()
		if err != nil {
			// Assume the branch doesn't exist and try to clone the default branch
			if b.branch != "" {
				// As of go1.5.1 linux/386, a Cmd struct can't be reused after calling its Run, Output or CombinedOutput methods.
				err := exec.Command(git, append(args[:len(args)-2], url, dest)[1:]...).Run()
				if err != nil {
					output.WriteString("can't be cloned!")
				} else {
					output.WriteString("cloned, but the branch " + b.branch + " doesn't exist.")
				}
			} else {
				output.WriteString("can't be cloned!")
			}
		} else {
			output.WriteString("cloned.")
		}
	} else if *update && headAttached(dest) {
		cmd.Dir = dest
		cmd.Args = strings.Fields("git pull")
		out, err := cmd.Output()
		if err != nil {
			output.WriteString(sep + url + " pull failed: " + err.Error())
		} else if len(out) != 0 && out[0] != 'A' { // out isn't "Already up-to-date"
			output.WriteString(sep + url + " updated.\n")
			log, _ := exec.Command(git, "-C", dest, "log", "--no-merges", "--oneline", "ORIG_HEAD..HEAD").Output()
			output.Write(bytes.TrimSpace(log))
			// Update submodules
			if _, err := os.Stat(dest + "/.gitmodules"); !os.IsNotExist(err) {
				exec.Command(git, "-C", dest, "submodule", "sync").Run()
				err := exec.Command(git, "-C", dest, "submodule", "update", "--init", "--recursive").Run()
				if err != nil {
					output.WriteString("\n------------ Submodule update failed: " + err.Error())
				}
			}
		}
	}
	if o := output.String(); len(o) != 0 {
		fmt.Println(o)
	}

}

// Check if git HEAD is in an attached state
func headAttached(path string) bool {
	f, err := os.Open(path + "/.git/HEAD")
	if err != nil {
		return true // assume attached
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if bytes.Contains(scanner.Bytes(), []byte("/")) {
			return true
		}
	}
	return false
}

// Clean removes disabled bundles from the file system
func Clean() {
	dirs, _ := filepath.Glob(root + "/*")
	var match bool
	for _, d := range dirs {
		match = false
		for _, b := range bundles {
			if d[strings.LastIndexAny(d, "/\\")+1:] == strings.Split(b.repo, "/")[1] {
				match = true
				break
			}
		}
		if !match {
			fmt.Print(sep)
			if *dryrun {
				fmt.Println(d, "would be removed.")
				return
			}
			var err error
			// Note: on Windows, read-only files woundn't be removed
			if runtime.GOOS == "windows" {
				err = exec.Command("cmd.exe", "/C", "rmdir", "/S", "/Q", d).Run()
			} else {
				err = os.RemoveAll(d)
			}
			if err != nil {
				fmt.Printf("Fail removing %v: %v\n", d, err)
			} else {
				fmt.Println(d, "removed.")
			}
		}
	}
}

// Bundles returns the bundle list output by Vim
func Bundles(bs ...string) []Bundle {
	vimrc := "NONE"
	for _, f := range []string{".vimrc", ".vim/vimrc", "_vimrc", "vimfiles/vimrc"} {
		ff := home + "/" + f
		_, err := os.Stat(ff)
		if os.IsNotExist(err) {
			continue
		} else {
			vimrc = ff
		}
	}
	args := []string{
		"-Nesu", vimrc,
		"--cmd", "let g:vundle = 1",
		"-c", "put =dundles | 2,print | quit!",
	}
	// there could be error even though the output is correct
	out, _ := exec.Command("vim", args...).Output()

	bundlesRaw := strings.Fields(string(out))
	bundles := make([]Bundle, len(bundlesRaw))
	for i, v := range bundlesRaw {
		bundles[i] = bundleDecode(v)
	}
	return bundles
}

// Decode a bundle of format: author/project[:[branch]][/sub/directory]
func bundleDecode(bi string) (bo Bundle) {
	var oneSlash bool
	// index the second slash
	slash2 := strings.IndexFunc(bi, func(r rune) bool {
		if r == '/' {
			if oneSlash == true {
				return true
			}
			oneSlash = true
		}
		return false
	})
	// remove [/sub/directory]
	if slash2 != -1 {
		bi = bi[:slash2]
	}
	bo = Bundle{bi, ""}
	// if [:[branch]] is present
	if bindex := strings.Index(bi, ":"); bindex >= 0 {
		bo.repo = (bi)[:bindex]
		if len(bi) == bindex+1 {
			bo.branch = runtime.GOOS + "_" + runtime.GOARCH
		} else {
			bo.branch = (bi)[bindex+1:]
		}
	}
	return
}

// Helptags generates Vim HELP tags for all bundles
func Helptags() {
	args := []string{
		"-Nesu",
		"NONE",
		"--cmd",
		`if &rtp !~# '\v[\/]\.vim[,|$]' | set rtp^=~/.vim | endif` +
			"| call rtp#inject() | Helptags" +
			func() string {
				if *update {
					return "!"
				}
				return ""
			}() + "| qall",
	}
	if exec.Command("vim", args...).Run() != nil {
		log.Printf("Fail generating HELP tags.")
	}
}

// vim:fdm=syntax:
