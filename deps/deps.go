// Package deps exists because shared libraries must be in the same directory
// as a go package for `go mod vendor` to work.
// (c.f. https://github.com/golang/go/issues/26366#issuecomment-405683150)
// The shared libraries are linked from result.go in the main package.
package deps
