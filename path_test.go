package automerge_test

import (
	"testing"

	"github.com/ConradIrwin/automerge-go"
	"github.com/stretchr/testify/require"
)

func TestPath_Errors(t *testing.T) {
	doc := automerge.New()

	require.PanicsWithError(t, "automerge: invalid path segment, expected string or int, got: bool(true)", func() {
		doc.Path(true)
	})

	_, err := doc.Path(0).Get()
	require.ErrorContains(t, err, "&automerge.Path{0}: tried to read index 0 of non-list &automerge.Map{}")

	err = doc.Path(0).Set(1)
	require.ErrorContains(t, err, "&automerge.Path{}: tried to write index 0 of non-list &automerge.Map{}")

	err = doc.Path("list", 1).Set(1)
	require.ErrorContains(t, err, `&automerge.Path{"list", 1}: tried to write index 1 beyond end of list length 0`)

	require.NoError(t, doc.Path("x").Set(true))

	_, err = doc.Path("x", "y").Get()
	require.ErrorContains(t, err, `&automerge.Path{"x", "y"}: tried to read property "y" of non-map true`)

	err = doc.Path("x", "y").Set(1)
	require.ErrorContains(t, err, `&automerge.Path{"x"}: tried to write property "y" of non-map true`)

	err = doc.Path().Set(1)
	require.ErrorContains(t, err, `&automerge.Path{}: tried to overwrite root of document`)
}
