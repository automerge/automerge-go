package automerge_test

import (
	"testing"

	"github.com/automerge/automerge-go"
	"github.com/stretchr/testify/require"
)

func TestStruct(t *testing.T) {
	type S struct {
		automerge.Struct

		Hello string `automerge:"hello"`
	}

	doc := automerge.New()
	require.NoError(t, doc.SetRoot(&S{Hello: "world"}))

	r, err := automerge.As[string](doc.Path("hello").Get())
	require.NoError(t, err)
	require.Equal(t, r, "world")

	require.NoError(t, doc.SetRoot((*S)(nil)))
}
