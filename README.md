automerge-go provides the ability to interact with [automerge] documents.
It is a featureful wrapper around [automerge-rs] that uses cgo to avoid reimplementing
the core engine from scratch.

## Installation

```
go get github.com/ConradIrwin/automerge-go
```

## Usage

See the Go documentation at https://pkg.go.dev/github.com/ConradIrwin/automerge-go.

## Limitations

automerge-go comes with precompiled libraries for linux and mac on amd64 (x86_64) and
arm64.

Supporting further architectures will require cross compiling automerge-rs and adding
them to the deps/ directory.

[automerge]: https://automerge.org
[automerge-rs]: https://github.com/automerge/automerge-rs
