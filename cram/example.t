  $ T="$(mktemp -d --suffix=".staged.cram")"
  $ trap "rm -rf -- \"$T\"" EXIT
  $ mkdir example
  $ cd example
  $ git init
  Initialized empty Git repository in .* (re)
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
