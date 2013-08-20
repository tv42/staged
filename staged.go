package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s COMMAND..\n", os.Args[0])
	flag.PrintDefaults()
}

func get_git_dir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		return "", fmt.Errorf("git directory looks wrong: %q", out)
	}
	return string(out[:len(out)-1]), nil
}

func get_toplevel() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		return "", fmt.Errorf("toplevel looks wrong: %q", out)
	}
	return string(out[:len(out)-1]), nil
}

// The subdirectory of the git worktree we are currently in, or empty
// string for root of worktree.
func get_git_prefix() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", nil
	}

	if out[len(out)-1] != '\n' {
		return "", fmt.Errorf("prefix looks wrong: %q", out)
	}
	return string(out[:len(out)-1]), nil
}

// Check whether the path is under a $GOPATH/src directory. If so,
// return the subtree path from there on. If not, return empty string.
func is_inside_gopath(p string) (string, error) {
	// i am butchering the difference between filepath and path.
	// i don't really care, right now.

	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("cannot make path absolute: %v", err)
	}
	abs = path.Clean(abs)

	for _, gopath := range filepath.SplitList(os.Getenv("GOPATH")) {
		gopath, err := filepath.Abs(gopath)
		if err != nil {
			return "", fmt.Errorf("cannot make path absolute: %v", err)
		}
		gopath = path.Clean(gopath)
		src := path.Join(gopath, "src") + "/"

		if strings.HasPrefix(abs, src) {
			rest := abs[len(src):]
			return rest, nil
		}
	}
	return "", nil
}

// Remove a variable from the environment. Returns the new
// environment.
func unsetenv(env []string, name string) []string {
	prefix := name + "="
	i := 0
	for {
		if i >= len(env) {
			break
		}
		for strings.HasPrefix(env[i], prefix) {
			if i+1 < len(env) {
				copy(env[i:], env[i+1:])
			}
			env = env[:len(env)-1]
			// don't increment i, look at this position again
			continue
		}
		i++
	}
	return env
}

func run(command string, args ...string) (err error) {
	gitdir, err := get_git_dir()
	if err != nil {
		return fmt.Errorf("cannot find git directory: %v", err)
	}
	gitdir, err = filepath.Abs(gitdir)
	if err != nil {
		return fmt.Errorf("cannot make git dir absolute: %v", err)
	}
	toplevel, err := get_toplevel()
	if err != nil {
		return fmt.Errorf("cannot find toplevel directory: %v", err)
	}
	prefix, err := get_git_prefix()
	if err != nil {
		return fmt.Errorf("cannot find worktree prefix: %v", err)
	}

	tmp, err := ioutil.TempDir(gitdir, "staged-")
	if err != nil {
		return fmt.Errorf("cannot create tempdir: %v", err)
	}
	defer func() {
		err2 := os.RemoveAll(tmp)
		if err2 != nil {
			// override the success with this cleanup error
			err = fmt.Errorf("tempdir cleanup failed: %v", err)
		}
	}()

	env := os.Environ()
	checkout_dir := tmp
	inside_gopath, err := is_inside_gopath(toplevel)
	if err != nil {
		return fmt.Errorf("cannot determine gopath: %v", err)
	}
	if inside_gopath != "" {
		gopath_list := []string{checkout_dir}
		gopath_list = append(gopath_list, filepath.SplitList(os.Getenv("GOPATH"))...)
		gopath := strings.Join(gopath_list, string(filepath.ListSeparator))

		// strip out existing GOPATH or we'll have duplicates
		// and undefined behavior
		env = unsetenv(env, "GOPATH")

		env = append(env, "GOPATH="+gopath)
		checkout_dir = path.Join(checkout_dir, "src", inside_gopath)
	}

	{
		cmd := exec.Command("git", "checkout-index", "--all", "--prefix="+checkout_dir+"/")
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("cannot checkout index: %v", err)
		}
	}

	{
		cmd := exec.Command(command, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// run in matching subdirectory; prefix may be empty
		// string, but Join handles that fine
		cmd.Dir = path.Join(checkout_dir, prefix)

		// sometimes we need to override GOPATH
		cmd.Env = env

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("command failed: %v", err)
		}
	}
	return nil
}

func main() {
	prog := path.Base(os.Args[0])
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = Usage
	flag.Parse()

	if flag.NArg() < 1 {
		Usage()
		os.Exit(1)
	}

	args := []string{}
	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}

	err := run(flag.Arg(0), args...)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
