package automerge_test

import (
	"fmt"
	"testing"

	"github.com/ConradIrwin/automerge-go"
	"github.com/stretchr/testify/require"
)

func TestActorId(t *testing.T) {
	a, err := automerge.NewActorId()
	require.NoError(t, err)

	b, err := automerge.ActorIdFromString(a.String())
	require.NoError(t, err)

	c, err := automerge.ActorIdFromBytes(b.Bytes())
	require.NoError(t, err)

	require.Equal(t, 0, a.Cmp(c))
	require.Equal(t, 0, b.Cmp(a))

	d, err := automerge.ActorIdFromString("x")
	require.Error(t, err)
	require.Equal(t, "Invalid actor ID: x", err.Error())
	require.Nil(t, d)

	// e, err := automerge.ActorIdFromBytes([]byte{})
	// require.Error(t, err)
	// require.Equal(t, "Invalid actor ID: x", err.Error())
	// require.Nil(t, e)

	f, err := automerge.ActorIdFromString("abcd")
	require.NoError(t, err)
	g, err := automerge.ActorIdFromString("cdef")
	require.NoError(t, err)

	require.Equal(t, -1, f.Cmp(g))
	require.Equal(t, 1, g.Cmp(f))
}

func TestDoc(t *testing.T) {

	d, err := automerge.New(nil)
	require.NoError(t, err)

	m, err := d.CreateMap("cards")
	require.NoError(t, err)
	fmt.Println(m)

}
