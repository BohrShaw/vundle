// Author: Bohr Shaw <pubohr@gmail.com>

// Vundle manages Vim bundles(plugins).
// It downloads, updates bundles, clean disabled bundles.
package main

import (
	"bufio"
	"bytes"
	"errors"
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
	Pull:
		pullCount := uint8(0)
		out, err := exec.Command(git, "-C", dest, "pull").Output()
		if err != nil {
			if pullCount < 3 {
				pullCount++
				goto Pull
			}
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
				continue
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

// Bundles returns the bundle list whose format is suitable for management.
func Bundles() []Bundle {
	rbundles := BundlesRaw()
	bundles := make([]Bundle, len(rbundles))
	i := 0
	for _, v := range rbundles {
		b, err := bundleDecode(v)
		if err != nil {
			fmt.Println(err)
			continue
		}
		bundles[i] = b
		i++
	}
	return bundles[:i]
}

// BundlesRaw returns the raw bundle list by parsing a specialized VimL file.
func BundlesRaw(files ...string) []string {
	file := home + "/.vim/init..vim"
	if files != nil {
		file = files[0]
	}
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	bundles := make([]string, 0, 100)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if regexp.MustCompile(`^\s*"`).Match(line) {
			continue
		}
		if i := bytes.Index(line, []byte("Bundle")); i >= 0 {
			j := i + 6
			if string(line[j:j+2]) == "s(" {
				if k := bytes.LastIndex(line, []byte(")")); k >= 0 {
					bundles = uappend(bundles, regexp.MustCompile(`[^ ,'"]+`).FindAllString(string(line[j+3:k-1]), -1)...)
				} else {
					log.Println("Arguments to Bundles() should be on a sigle line.")
				}
			} else if i := regexp.MustCompile(`^\w*\(`).FindIndex(line[j:]); i != nil {
				bundles = uappend(bundles, regexp.MustCompile(`[^ '"]+`).FindString(string(line[j+i[1]+1:])))
			}
		}
	}
	return bundles
}

// Decode a bundle of format: author/project[:[branch]][/sub/directory]
func bundleDecode(bi string) (bo Bundle, _ error) {
	bundleFormat := regexp.MustCompile(
		`^[[:word:]-.]+/[[:word:]-.]+` + // author/project
			`(:([[:word:]-.]+)?)?` + // [:[branch]]
			`([[:word:]-.]+/)*[[:word:]-.]*$`) // [/sub/directory]
	if !bundleFormat.MatchString(bi) {
		return bo, errors.New("Wrong bundle format: " + bi)
	}
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
	return bo, nil
}

// uappend appends elements to a slice whose elements remains distinct,
// ignoring case sensitivity.
func uappend(slice []string, elems ...string) (result []string) {
	result = slice
	for _, e := range elems {
		e = strings.ToLower(e)
		equal := false
		for _, s := range slice {
			s = strings.ToLower(s)
			if s == e {
				equal = true
				break
			}
		}
		if !equal {
			result = append(result, e)
		}
	}
	return
}

// Helptags generates Vim HELP tags for all bundles
func Helptags() {
	overwrite := "0"
	if *update {
		overwrite = "1"
	}
	args := []string{
		"-Nes", "--cmd",
		"set rtp^=~/.vim | call helptags#(" + overwrite + ") | qall!",
	}
	vim, err := exec.LookPath("vim")
	if err != nil {
		vim, err = exec.LookPath("nvim")
		if err != nil {
			vim, err = exec.LookPath("gvim")
		}
	}
	if exec.Command(vim, args...).Run() != nil {
		log.Printf("Fail generating HELP tags.")
	}
}

// vim:fdm=syntax:
