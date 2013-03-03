staged -- Run a command with the Git staged files
=================================================

Put contents of Git index in a temporary directory and run a command
there. This makes unit testing "git add -p" results easy, and helps
you avoid the common problem where you forget to commit changes in
another file.

Install with

    go get github.com/tv42/staged

Use like this:

    $ mkdir example
    $ cd example
    $ git init
    $ echo first line >one
    $ echo noise >two
    $ git add one
    $ staged ls
    one
    $ echo more >>one
    $ cat one
    first line
    more
    $ staged cat one
    first line
