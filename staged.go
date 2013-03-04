package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s COMMAND..\n", os.Args[0])
	flag.PrintDefaults()
}

func get_git_dir() string {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("cannot find git directory")
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		log.Fatalf("git directory looks wrong: %q", out)
	}
	return string(out[:len(out)-1])
}

// The subdirectory of the git worktree we are currently in, or empty
// string for root of worktree.
func get_git_prefix() string {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("cannot find worktree prefix")
	}
	if len(out) == 0 {
		return ""
	}

	if out[len(out)-1] != '\n' {
		log.Fatalf("prefix looks wrong: %q", out)
	}
	return string(out[:len(out)-1])
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

	gitdir := get_git_dir()
	prefix := get_git_prefix()

	tmp, err := ioutil.TempDir(gitdir, "staged-")
	if err != nil {
		log.Fatalf("cannot create tempdir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			log.Fatalf("tempdir cleanup failed: %v", err)
		}
	}()

	{
		cmd := exec.Command("git", "checkout-index", "--all", "--prefix="+tmp+"/")
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("cannot checkout index: %v", err)
		}
	}

	{
		args := []string{}
		if flag.NArg() > 1 {
			args = flag.Args()[1:]
		}
		cmd := exec.Command(flag.Arg(0), args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// run in matching subdirectory; prefix may be empty
		// string, but Join handles that fine
		cmd.Dir = path.Join(tmp, prefix)

		err = cmd.Run()
		if err != nil {
			log.Fatalf("command failed: %v", err)
		}
	}
}
