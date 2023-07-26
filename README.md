# automerge-go

automerge-go provides the ability to interact with [automerge] documents.
It is a featureful wrapper around [automerge-rs] that uses cgo to avoid reimplementing
the core engine from scratch.

For package documentation, see the Go documentation at https://pkg.go.dev/github.com/automerge/automerge-go.

[automerge]: https://automerge.org
[automerge-rs]: https://github.com/automerge/automerge-rs

## Building from automerge-c
This step must be done on an Apple Silicon mac with docker running.

```sh
./deps/rebuild.sh
```