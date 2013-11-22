# -*- indent-tabs-mode: nil -*-

  $ export STAGED_CACHE_DIR="$PWD/staged-cache"
  $ # step 1 OMIT
  $ git init --quiet
  $ cat >demo.go <<'EOF'
  > package demo
  > func foo() int {
  >     return 42
  > }
  > EOF
  $ cat >demo_test.go <<'EOF'
  > package demo
  > import "testing"
  > func TestFoo(t *testing.T) {
  >     if g, e := foo(), 42; g != e {
  >         t.Errorf("bad foo: %v != %v", g, e)
  >     }
  > }
  > EOF
  $ # step 2 OMIT
  $ git add demo_test.go        # forget to add demo.go
  $ go test                     # ignores the problem
  PASS
  ok * (glob)
  $ staged go test              # detects the problem
  # * (glob)
  ./demo_test.go:4: undefined: foo
  FAIL\t* [build failed] (esc) (glob)
  staged: command failed: exit status 2
  [1]
  $ staged ls                   # no demo.go
  demo_test.go
  $ git add demo.go
  $ staged go test
  PASS
  ok * (glob)
