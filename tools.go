//go:build skiptools

/*
This file keeps versions of tools needed to develop automerge-go

To lint, `go run honnef.co/go/tools/cmd/staticcheck` (or
`go install honnef.co/go/tools/cmd/staticcheck` and then just
run `staticcheck`)
*/
package automerge

import _ "honnef.co/go/tools/cmd/staticcheck"
