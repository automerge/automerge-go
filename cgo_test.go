package automerge_test

import (
	"fmt"
	"testing"
	"time"

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

	doc, err := automerge.New(a)
	require.NoError(t, err)
	a2, err := doc.ActorId()
	require.NoError(t, err)
	require.Equal(t, a.String(), a2.String())
}

func TestDoc(t *testing.T) {

	d, err := automerge.New(nil)
	require.NoError(t, err)

	err = d.Root().Set("x", "bloop")
	require.NoError(t, err)

	b, err := d.Save()
	require.NoError(t, err)

	d, err = automerge.Load(b)
	require.NoError(t, err)

	s, err := automerge.As[string](d.Root().Get("x"))
	require.NoError(t, err)
	require.Equal(t, s, "bloop")

	// from @automerge/automerge 2.0.0-beta.4
	// var newDoc = automerge.change(doc, (d: any) => {
	// 	d.bool = true
	// 	d.string = 'hello'
	// 	d.int = new automerge.Uint(4)
	// 	d.uint = new automerge.Uint(4)
	// 	d.float = 3.14
	// 	d.time = new Date('2000-01-02T03:04:05.006Z')
	// 	d.bytes = new Uint8Array([1, 0, 2, 0, 3])
	// 	d.slice = ['a', 'b', 'c']
	// 	d.map = { a: 1, b: 2 }
	// 	d.counter = new automerge.Counter(10)
	// 	d.text = new automerge.Text('hello world')
	// })
	bytes := []byte{133, 111, 74, 131, 168, 144, 130, 162, 0, 180, 2, 1, 16, 255, 24, 175, 194, 49, 172, 71, 83, 132, 87, 37, 154, 215, 119, 247, 226, 1, 195, 106, 236, 214, 77, 0, 106, 11, 204, 159, 199, 153, 79, 29, 24, 247, 79, 172, 12, 56, 245, 104, 101, 39, 77, 250, 104, 236, 129, 166, 17, 219, 6, 1, 2, 3, 2, 19, 2, 35, 2, 64, 2, 86, 2, 12, 1, 4, 2, 8, 17, 8, 19, 13, 21, 71, 33, 2, 35, 21, 52, 4, 66, 9, 86, 20, 87, 43, 128, 1, 2, 127, 0, 127, 1, 127, 27, 127, 0, 127, 0, 127, 7, 0, 11, 16, 0, 0, 11, 3, 8, 2, 12, 11, 16, 0, 12, 2, 0, 0, 3, 10, 0, 0, 11, 125, 0, 9, 1, 0, 2, 126, 118, 17, 9, 1, 117, 4, 98, 111, 111, 108, 5, 98, 121, 116, 101, 115, 7, 99, 111, 117, 110, 116, 101, 114, 5, 102, 108, 111, 97, 116, 3, 105, 110, 116, 3, 109, 97, 112, 5, 115, 108, 105, 99, 101, 6, 115, 116, 114, 105, 110, 103, 4, 116, 101, 120, 116, 4, 116, 105, 109, 101, 4, 117, 105, 110, 116, 0, 3, 126, 1, 97, 1, 98, 0, 11, 27, 0, 116, 1, 6, 8, 118, 126, 9, 124, 122, 14, 118, 126, 5, 2, 1, 125, 2, 1, 3, 10, 1, 11, 3, 2, 11, 5, 1, 124, 0, 2, 1, 4, 18, 1, 123, 2, 87, 24, 133, 1, 19, 2, 0, 124, 86, 0, 105, 19, 3, 22, 2, 20, 11, 22, 1, 0, 2, 0, 3, 10, 31, 133, 235, 81, 184, 30, 9, 64, 4, 104, 101, 108, 108, 111, 142, 161, 250, 132, 199, 27, 4, 97, 98, 99, 1, 2, 104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 27, 0, 0}
	d, err = automerge.Load(bytes)
	require.NoError(t, err)

	m := d.Root()

	testGet(t, m, "bool", true)
	testGet(t, m, "string", "hello")
	testGet(t, m, "int", int(4))
	testGet(t, m, "uint", int(4))
	testGet(t, m, "float", 3.14)
	ts, err := time.Parse(time.RFC3339Nano, "2000-01-02T03:04:05.006Z")
	require.NoError(t, err)
	testGet(t, m, "time", time.UnixMilli(ts.UnixMilli()))
	testGet(t, m, "bytes", []byte{1, 0, 2, 0, 3})
	testGet(t, m, "slice", []any{"a", "b", "c"})
	testGet(t, m, "map", map[string]any{"a": 1.0, "b": 2.0})
	testGet(t, m, "counter", 10)
	testGet(t, m, "text", "hello world")

	c, err := automerge.As[*automerge.Counter](m.Get("counter"))
	require.NoError(t, err)
	n, err := c.Get()
	require.NoError(t, err)
	require.Equal(t, int64(10), n)

	txt, err := automerge.As[*automerge.Text](m.Get("text"))
	require.NoError(t, err)
	s, err = txt.Get()
	require.NoError(t, err)
	require.Equal(t, s, "hello world")
}

func testGet[T any](t *testing.T, m *automerge.Map, k string, v T) {
	got, err := automerge.As[T](m.Get(k))
	require.NoError(t, err)
	require.Equal(t, v, got)
}

func TestMap(t *testing.T) {
	doc, err := automerge.New(nil)
	require.NoError(t, err)

	// create new map
	m := automerge.NewMap()
	doc.Root().Set("x", m)

	fmt.Printf("%#v", m)

	now := time.Now().Round(time.Millisecond)

	require.NoError(t, m.Set("bool", true))
	require.NoError(t, m.Set("string", "hello"))
	require.NoError(t, m.Set("bytes", []byte{10, 0, 10, 0, 10}))
	require.NoError(t, m.Set("int", 4))
	require.NoError(t, m.Set("uint", uint(4)))
	require.NoError(t, m.Set("float", 3.14))
	require.NoError(t, m.Set("time", now))
	require.NoError(t, m.Set("slice", []string{"a", "b", "c"}))
	require.NoError(t, m.Set("map", map[string]int{"a": 1, "b": 2}))

	testGet(t, m, "bool", true)
	testGet(t, m, "string", "hello")
	testGet(t, m, "bytes", []byte{10, 0, 10, 0, 10})
	testGet(t, m, "int", 4)
	testGet[uint](t, m, "uint", 4)
	testGet(t, m, "float", 3.14)
	testGet(t, m, "time", now)
	testGet(t, m, "slice", []string{"a", "b", "c"})
	testGet(t, m, "map", map[string]int{"a": 1, "b": 2})

	require.NoError(t, m.Del("map"))
	v, err := m.Get("map")
	require.NoError(t, err)
	require.True(t, v.IsVoid())
}

func TestMap_Path(t *testing.T) {
	doc, err := automerge.New(nil)
	require.NoError(t, err)

	i := doc.Path("x").Map().Iter()
	for {
		_, _, valid := i.Next()
		if !valid {
			break
		}
		t.Fatal("expected non-extant map to have no items")
	}
	require.NoError(t, i.Error())

	v, err := doc.Path("x").Map().Get("y")
	require.NoError(t, err)
	require.True(t, v.IsVoid())

	v, err = doc.Path("y", 0).Map().Get("y")
	require.NoError(t, err)
	require.True(t, v.IsVoid())

	err = doc.Path("x").Map().Set("y", true)
	require.NoError(t, err)

	b, err := automerge.As[bool](doc.Path("x").Map().Get("y"))
	require.NoError(t, err)
	require.True(t, b)

	err = doc.Path("y", 0).Map().Set("y", true)
	require.NoError(t, err)

	b, err = automerge.As[bool](doc.Path("y", 0).Map().Get("y"))
	require.NoError(t, err)
	require.True(t, b)
}

func TestCounter(t *testing.T) {
	doc, err := automerge.New(nil)
	require.NoError(t, err)

	c := automerge.NewCounter(10)
	require.NoError(t, doc.Root().Set("x", c))
	require.NoError(t, c.Inc(-20))

	b := automerge.NewCounter(5)
	require.NoError(t, doc.Path("y", 0).Put(b))
	require.NoError(t, b.Inc(20))

	v, err := automerge.As[int8](doc.Root().Get("x"))
	require.NoError(t, err)
	fmt.Println("got", v)

	v, err = automerge.As[int8](doc.Path("y", 0).Get())
	require.NoError(t, err)
	fmt.Println("got", v)
}

func TestText(t *testing.T) {
	doc, err := automerge.New(nil)
	require.NoError(t, err)

	txt := automerge.NewText("hello world")
	require.NoError(t, doc.Root().Set("x", txt))

	doc2, err := doc.Fork(nil)
	require.NoError(t, err)

	txt2, err := automerge.As[*automerge.Text](doc2.Root().Get("x"))
	require.NoError(t, err)

	require.NoError(t, txt.Append("!"))
	require.NoError(t, txt2.Insert(6, "cool "))

	_, err = doc.Merge(doc2)
	require.NoError(t, err)

	v, err := automerge.As[string](doc.Root().Get("x"))
	require.Equal(t, "hello cool world!", v)
}

/*
func FuzzLoad(f *testing.F) {
	testcases := [][]byte{
		[]byte{},
		[]byte{133, 111, 74, 131, 68, 61, 163, 83, 0, 113, 1, 16, 91, 26, 133, 110, 147, 199, 78, 47, 163, 33, 169, 228, 102, 47, 254, 186, 1, 59, 30, 223, 49, 84, 152, 119, 143, 189, 251, 117, 46, 67, 252, 22, 238, 237, 195, 20, 27, 172, 151, 163, 65, 65, 211, 66, 231, 54, 58, 98, 167, 6, 1, 2, 3, 2, 19, 2, 35, 2, 64, 2, 86, 2, 8, 21, 4, 33, 2, 35, 2, 52, 1, 66, 2, 86, 2, 87, 3, 128, 1, 2, 127, 0, 127, 1, 127, 1, 127, 0, 127, 0, 127, 7, 127, 2, 102, 122, 127, 0, 127, 1, 1, 127, 1, 127, 54, 239, 191, 189, 127, 0, 0},
		// []byte{133, 111, 75, 131, 56, 24, 10, 149, 0, 116, 1, 16, 2, 81, 166, 112, 192, 209, 78, 82, 128, 106, 6, 204, 35, 117, 116, 163, 1, 183, 132, 110, 47, 155, 92, 108, 170, 222, 200, 214, 90, 248, 97, 205, 41, 48, 73, 30, 136, 101, 35, 40, 144, 29, 82, 172, 124, 168, 204, 178, 228, 6, 1, 2, 3, 2, 19, 2, 35, 2, 64, 2, 86, 2, 8, 21, 6, 33, 2, 35, 2, 52, 1, 66, 2, 86, 2, 87, 4, 128, 1, 2, 127, 0, 127, 1, 127, 1, 127, 0, 127, 0, 127, 7, 127, 4, 111, 111, 112, 115, 127, 0, 127, 1, 1, 127, 1, 127, 70, 111, 0, 112, 115, 127, 0, 0},
		// []byte{133, 111, 74, 131, 116, 189, 7, 51, 0, 116, 1, 16, 239, 176, 188, 84, 137, 139, 66, 210, 143, 133, 247, 153, 112, 114, 20, 232, 1, 152, 25, 208, 42, 123, 119, 195, 139, 165, 44, 212, 240, 235, 71, 33, 253, 198, 218, 25, 2, 249, 44, 66, 3, 150, 107, 93, 40, 243, 89, 199, 155, 6, 1, 2, 3, 2, 19, 2, 35, 2, 64, 2, 86, 2, 8, 21, 6, 33, 2, 35, 2, 52, 1, 66, 2, 86, 2, 87, 4, 128, 1, 2, 127, 0, 127, 1, 127, 1, 127, 0, 127, 0, 127, 7, 127, 4, 111, 0, 112, 115, 127, 0, 127, 1, 1, 127, 1, 127, 70, 111, 111, 112, 115, 127, 0, 0},
		[]byte{133, 111, 74, 131, 166, 255, 147, 49, 0, 114, 1, 16, 62, 147, 195, 223, 55, 54, 72, 60, 138, 182, 74, 179, 74, 23, 56, 204, 1, 199, 197, 206, 98, 174, 2, 51, 252, 197, 139, 198, 34, 52, 107, 160, 19, 168, 20, 101, 184, 247, 160, 123, 210, 222, 245, 191, 211, 109, 235, 212, 38, 6, 1, 2, 3, 2, 19, 2, 35, 2, 64, 2, 86, 2, 8, 21, 4, 33, 2, 35, 2, 52, 1, 66, 2, 86, 2, 87, 4, 128, 1, 2, 127, 0, 127, 1, 127, 1, 127, 0, 127, 0, 127, 7, 127, 2, 102, 122, 127, 0, 127, 1, 1, 127, 1, 127, 70, 240, 159, 167, 159, 127, 0, 0},
	}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig []byte) {
		d, err := automerge.Load(orig)
		if err != nil {
			return
		}

		automerge.As[map[string]any](d.Get())
	})
}
*/
