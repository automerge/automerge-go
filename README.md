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

Currently this only works on macOS with an M-series processor, adding more architectures
will require cross-compiling [automerge-rs].

[automerge]: https://automerge.org
[automerge-rs]: https://github.com/automerge/automerge-rs
