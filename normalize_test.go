package automerge_test

import (
	"testing"

	"github.com/automerge/automerge-go"
	"github.com/stretchr/testify/require"
)

func TestAs(t *testing.T) {

	doc := automerge.New()
	require.NoError(t, doc.RootMap().Set("float64", float64(1.5)))
	l := automerge.NewList()
	require.NoError(t, doc.RootMap().Set("list", l))
	require.NoError(t, l.Append(1))

	ret, err := automerge.As[any](doc.RootMap().Get("list"))
	require.NoError(t, err)
	require.Equal(t, ret, []any{float64(1)})

	rets, err := automerge.As[[]any](doc.RootMap().Get("list"))
	require.NoError(t, err)
	require.Equal(t, rets, []any{float64(1)})

	require.NoError(t, doc.Path("list").List().Append(2))
	arr, err := automerge.As[[2]uint16](doc.RootMap().Get("list"))
	require.NoError(t, err)
	require.Equal(t, arr, [2]uint16{1, 2})

	m := automerge.NewMap()
	require.NoError(t, doc.RootMap().Set("map", m))
	require.NoError(t, m.Set("x", 5))

	d, err := automerge.As[float64](doc.RootMap().Get("float64"))
	require.NoError(t, err)
	require.Equal(t, d, 1.5)

	ret2, err := automerge.As[[]int](doc.RootMap().Get("list"))
	require.NoError(t, err)
	require.Equal(t, ret2, []int{1, 2})

	retm, err := automerge.As[map[string]int](doc.RootMap().Get("map"))
	require.NoError(t, err)
	require.Equal(t, retm, map[string]int{"x": 5})

	require.NoError(t, m.Set("one", "six"))
	require.NoError(t, m.Set("two", uint(5)))

	type Y string
	type X struct {
		X int `automerge:"x"`
		Y Y   `automerge:"one"`
	}

	r, err := automerge.As[*X](doc.RootMap().Get("map"))
	require.NoError(t, err)
	require.Equal(t, r, &X{5, Y("six")})

	require.Error(t, doc.RootMap().Set("map", map[int]int{1: 1}))
}

func TestTags(t *testing.T) {
	doc := automerge.New()
	require.Error(t, doc.RootMap().Set("map", map[int]int{1: 1}))

	type X struct {
		X int     `automerge:"x"`
		Y uint    `automerge:"-"`
		Z float32 `automerge:"z"`
	}

	require.NoError(t, doc.RootMap().Set("x", &X{1, 2, 0}))

	m, err := automerge.As[map[string]any](doc.RootMap().Get("x"))
	require.NoError(t, err)
	require.Equal(t, m, map[string]any{"x": 1.0, "z": 0.0})

	type S struct {
		H string `automerge:"h"`
		i int64  `automerge:"i"`
	}

	require.NoError(t, doc.RootMap().Set("s", &S{H: "hello", i: 5}))
	require.NoError(t, doc.Path("s", "i").Set(7))

	s, err := automerge.As[S](doc.Path("s").Get())
	require.NoError(t, err)
	require.Equal(t, s.H, "hello")
	require.Equal(t, s.i, int64(0))

	v, err := doc.Path("i").Get()
	require.NoError(t, err)
	require.True(t, v.IsVoid())
}
