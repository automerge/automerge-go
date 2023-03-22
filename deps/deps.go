// Package deps exists because shared libraries must be in the same directory
// as a go package for `go mod vendor` to work.
// (c.f. https://github.com/golang/go/issues/26366#issuecomment-405683150)
package deps

/*
#cgo LDFLAGS: -L${SRCDIR}
#cgo darwin,arm64 LDFLAGS: -lautomerge_core_darwin_arm64
#cgo darwin,amd64 LDFLAGS: -lautomerge_core_darwin_amd64
#cgo linux,arm64 LDFLAGS: -lautomerge_core_linux_arm64 -lm
#cgo linux,amd64 LDFLAGS: -lautomerge_core_linux_amd64 -lm
*/
import "C"
