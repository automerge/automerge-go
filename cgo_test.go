package automerge_test

import (
	"encoding/base64"
	"runtime"
	"testing"
	"time"

	"github.com/automerge/automerge-go"
	"github.com/stretchr/testify/require"
)

func TestActorID(t *testing.T) {
	a := automerge.NewActorID()

	b, err := automerge.ActorIDFromString(a.String())
	require.NoError(t, err)

	c, err := automerge.ActorIDFromBytes(b.Bytes())
	require.NoError(t, err)

	require.Equal(t, 0, a.Cmp(c))
	require.Equal(t, 0, b.Cmp(a))

	d, err := automerge.ActorIDFromString("x")
	require.Error(t, err)
	require.Contains(t, err.Error(), "nvalid actor ID")
	require.Nil(t, d)

	// e, err := automerge.ActorIDFromBytes([]byte{})
	// require.Error(t, err)
	// require.Equal(t, "Invalid actor ID: x", err.Error())
	// require.Nil(t, e)

	f, err := automerge.ActorIDFromString("abcd")
	require.NoError(t, err)
	g, err := automerge.ActorIDFromString("cdef")
	require.NoError(t, err)

	require.Equal(t, -1, f.Cmp(g))
	require.Equal(t, 1, g.Cmp(f))

	doc := automerge.New()
	require.NoError(t, doc.SetActorID(f))
	f2, err := doc.ActorID()
	require.NoError(t, err)
	require.Equal(t, f.String(), f2.String())
}

func TestDoc(t *testing.T) {
	d := automerge.New()

	err := d.RootMap().Set("x", "bloop")
	require.NoError(t, err)

	b, err := d.Save()
	require.NoError(t, err)

	d, err = automerge.Load(b)
	require.NoError(t, err)

	s, err := automerge.As[string](d.RootMap().Get("x"))
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

	m := d.RootMap()

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
	testGet(t, m, "map", map[string]any{"a": int64(1), "b": int64(2)})
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

func TestDoc_Fork(t *testing.T) {
	d := automerge.New()
	require.NoError(t, d.RootMap().Set("x", 1))

	ch, err := d.Commit("initial version")
	require.NoError(t, err)

	ch2, err := d.Heads()
	require.NoError(t, err)
	require.Equal(t, 0, ch.Cmp(ch2))

	require.Equal(t, ch.Get()[0].String(), ch2.Get()[0].String())

	require.NoError(t, d.RootMap().Set("x", 2))

	d2, err := d.Fork(ch)
	require.NoError(t, err)

	v, err := automerge.As[int](d.RootMap().Get("x"))
	require.NoError(t, err)
	v2, err := automerge.As[int](d2.RootMap().Get("x"))
	require.NoError(t, err)

	require.Equal(t, v, 2)
	require.Equal(t, v2, 1)
}

func TestDoc_Errors(t *testing.T) {
	ai, err := automerge.ActorIDFromString("5a1aac51ffc84d6cb7b72c626b35f962")
	require.NoError(t, err)
	d := automerge.New()
	d2 := automerge.New()
	d.SetActorID(ai)
	d2.SetActorID(ai)

	require.NoError(t, d.RootMap().Set("x", map[string]any{"x": 1}))
	require.NoError(t, d2.RootMap().Set("y", map[string]any{"y": 1}))

	_, err = d.Merge(d2)
	require.EqualError(t, err, "duplicate seq 1 found for actor 5a1aac51ffc84d6cb7b72c626b35f962")
}

func testGet[T any](t *testing.T, m *automerge.Map, k string, v T) {
	got, err := automerge.As[T](m.Get(k))
	require.NoError(t, err)
	require.Equal(t, v, got)
}

func TestMap(t *testing.T) {
	doc := automerge.New()

	// create new map
	m := automerge.NewMap()
	doc.RootMap().Set("x", m)

	now := time.Now().Round(time.Millisecond)

	require.NoError(t, m.Set("bool", true))
	require.NoError(t, m.Set("string", "hello"))
	require.NoError(t, m.Set("bytes", []byte{10, 0, 10, 0, 10}))
	require.NoError(t, m.Set("int", int64(4)))
	require.NoError(t, m.Set("uint", uint64(4)))
	require.NoError(t, m.Set("float", 3.14))
	require.NoError(t, m.Set("time", now))
	require.NoError(t, m.Set("null", nil))
	require.NoError(t, m.Set("slice", []string{"a", "b", "c"}))
	require.NoError(t, m.Set("map", map[string]int{"a": 1, "b": 2}))

	require.Equal(t, "&automerge.Map{\"bool\": true, \"bytes\": []byte{0xa, 0x0, 0xa, 0x0, 0xa}, \"float\": 3.14, \"int\": 4, \"map\": &automerge.Map{...}, ...}", m.GoString())

	testGet(t, m, "bool", true)
	testGet(t, m, "string", "hello")
	testGet(t, m, "bytes", []byte{10, 0, 10, 0, 10})
	testGet(t, m, "int", 4)
	testGet[uint](t, m, "uint", 4)
	testGet(t, m, "float", 3.14)
	testGet(t, m, "time", now)
	testGet[any](t, m, "null", nil)
	testGet(t, m, "slice", []string{"a", "b", "c"})
	testGet(t, m, "map", map[string]int{"a": 1, "b": 2})

	require.NoError(t, m.Del("map"))
	v, err := m.Get("map")
	require.NoError(t, err)
	require.True(t, v.IsVoid())
}

func TestLoad(t *testing.T) {
	/*
		import * as automerge from '@automerge/automerge' // 2.0.0-beta.4
		let doc = automerge.init()
		var newDoc = automerge.change(doc, (d: any) => {
			d.oops = 'o\x00ps'
		})
		console.log(Buffer.from(automerge.save(newDoc)).toString('base64'))
	*/
	withNullV := "hW9KgzgYCpUAdAEQAlGmcMDRTlKAagbMI3V0owG3hG4vm1xsqt7I1lr4Yc0pMEkeiGUjKJAdUqx8qMyy5AYBAgMCEwIjAkACVgIIFQYhAiMCNAFCAlYCVwSAAQJ/AH8BfwF/AH8Afwd/BG9vcHN/AH8BAX8Bf0ZvAHBzfwAA"
	b, err := base64.StdEncoding.DecodeString(withNullV)
	require.NoError(t, err)

	doc, err := automerge.Load(b)
	require.NoError(t, err)

	s, err := automerge.As[string](doc.Path("oops").Get())
	require.NoError(t, err)
	require.Equal(t, "o\x00ps", s)

	/*
		import * as automerge from '@automerge/automerge'>
		let doc = automerge.init()
		var newDoc = automerge.change(doc, (d: any) => {
			d['o\x00ps'] = 'oops'
		})
		console.log(Buffer.from(automerge.save(newDoc)).toString('base64'))
	*/
	withNullK := "hW9Kgw0rK0gAdAEQiIyyxW8dRnun78aXDbEXJAEVxQ+rJJGUfhMX00tXOBpup2Mg7zMrjQojGW0d1PelNgYBAgMCEwIjAkACVgIIFQYhAiMCNAFCAlYCVwSAAQJ/AH8BfwF/AH8Afwd/BG8AcHN/AH8BAX8Bf0Zvb3BzfwAA"
	b, err = base64.StdEncoding.DecodeString(withNullK)
	require.NoError(t, err)

	doc, err = automerge.Load(b)
	require.NoError(t, err)

	iter := doc.RootMap().Iter()
	for {
		k, v, valid := iter.Next()
		if !valid {
			break
		}
		require.Equal(t, "o\x00ps", k)
		require.Equal(t, "oops", v.Str())
	}
	require.NoError(t, iter.Error())
}

func TestMap_Path(t *testing.T) {
	doc := automerge.New()

	i := doc.Path("x").Map().Iter()
	for {
		_, _, valid := i.Next()
		if !valid {
			break
		}
		t.Fatal("expected non-extant map to have no items")
	}
	require.NoError(t, i.Error())

	require.Equal(t, 0, doc.Path("x").Map().Len())

	v, err := doc.Path("x").Map().Get("y")
	require.NoError(t, err)
	require.True(t, v.IsVoid())

	v, err = doc.Path("y", 0).Map().Get("y")
	require.NoError(t, err)
	require.True(t, v.IsVoid())

	err = doc.Path("x").Map().Set("y", true)
	require.NoError(t, err)
	err = doc.Path("x").Map().Set("z", true)
	require.NoError(t, err)

	require.Equal(t, 2, doc.Path("x").Map().Len())

	b, err := automerge.As[bool](doc.Path("x").Map().Get("y"))
	require.NoError(t, err)
	require.True(t, b)

	err = doc.Path("y", 0).Map().Set("y", true)
	require.NoError(t, err)

	b, err = automerge.As[bool](doc.Path("y", 0).Map().Get("y"))
	require.NoError(t, err)
	require.True(t, b)

	iter := doc.Path("y", 0).Map().Iter()
	for {
		k, v, valid := iter.Next()
		if !valid {
			break
		}
		require.Equal(t, "y", k)
		require.Equal(t, true, v.Bool())

	}
	require.NoError(t, iter.Error())

	_, _, valid := iter.Next()
	require.False(t, valid)
	require.NoError(t, iter.Error())

	iter = doc.Path("y").Map().Iter()
	for {
		_, _, valid := iter.Next()
		if !valid {
			break
		}
		t.Fatalf("expected no valid iterations")
	}
	require.EqualError(t, iter.Error(), "&automerge.Path{\"y\"}: tried to interate over non-map &automerge.List{&automerge.Map{...}}")
}

func TestList(t *testing.T) {
	doc := automerge.New()

	l := automerge.NewList()
	require.NoError(t, doc.RootMap().Set("l", l))

	require.NoError(t, l.Append("a", "b", "e"))
	require.NoError(t, l.Insert(2, "d", "d", "d"))
	require.NoError(t, l.Set(2, "c"))
	require.NoError(t, l.Delete(3))

	v, err := automerge.As[[]string](doc.RootMap().Get("l"))
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, v)

	require.EqualError(t, l.Set(10, "d"), "automerge.List: tried to write index 10 beyond end of list length 5")

	require.NoError(t, l.Insert(5, "f"))

	require.Equal(t, `&automerge.List{"a", "b", "c", "d", "e", ...}`, l.GoString())

	m := automerge.NewList()
	require.EqualError(t, m.Append(nil), "automerge.List: tried to write to detached list")
	require.NoError(t, doc.RootMap().Set("m", m))

	now := time.UnixMilli(time.Now().UnixMilli())

	require.NoError(t, m.Append(nil))
	require.NoError(t, m.Append(true))
	require.NoError(t, m.Append(1))
	require.NoError(t, m.Append(int64(2)))
	require.NoError(t, m.Append(uint64(4611686018427387904)))
	require.NoError(t, m.Append("hello"))
	require.NoError(t, m.Append([]byte{0, 10, 0}))
	require.NoError(t, m.Append(now))
	require.NoError(t, m.Append([]string{"a", "b", "c"}))
	require.NoError(t, m.Append(map[string]string{"🧟": "🛩"}))
	require.NoError(t, m.Append(automerge.NewCounter(1)))
	require.NoError(t, m.Append(automerge.NewText("a")))
	require.NoError(t, m.Append(automerge.NewMap()))
	require.NoError(t, m.Append(automerge.NewList()))

	out, err := automerge.As[[]any](doc.RootMap().Get("m"))
	require.NoError(t, err)
	require.Equal(t, []any{nil, true,
		float64(1), int64(2), uint64(4611686018427387904),
		"hello", []byte{0, 10, 0}, now,
		[]any{"a", "b", "c"}, map[string]any{"🧟": "🛩"},
		int64(1), "a", map[string]any{}, []any{}}, out)

	bytes, err := doc.Save()
	require.NoError(t, err)

	doc, err = automerge.Load(bytes)
	require.NoError(t, err)
	val, err := doc.RootMap().Get("m")
	require.NoError(t, err)
	m = val.List()

	mustGet := func(i int) *automerge.Value {
		v, err := m.Get(i)
		require.NoError(t, err)
		return v
	}
	require.True(t, mustGet(0).IsNull())
	require.True(t, mustGet(1).Bool())
	require.Equal(t, 1.0, mustGet(2).Float64())
	require.Equal(t, int64(2), mustGet(3).Int64())
	require.Equal(t, uint64(4611686018427387904), mustGet(4).Uint64())
	require.Equal(t, "hello", mustGet(5).Str())
	require.Equal(t, []byte{0, 10, 0}, mustGet(6).Bytes())
	require.Equal(t, now, mustGet(7).Time())

	l = mustGet(8).List()
	require.Equal(t, 3, l.Len())

	mp := mustGet(9).Map()
	s, err := automerge.As[string](mp.Get("🧟"))
	require.NoError(t, err)
	require.Equal(t, "🛩", s)

	c := mustGet(10).Counter()
	cnt, err := c.Get()
	require.NoError(t, err)
	require.Equal(t, int64(1), cnt)

	txt := mustGet(11).Text()
	s, err = txt.Get()
	require.NoError(t, err)
	require.Equal(t, "a", s)

	mp = mustGet(12).Map()
	require.Equal(t, 0, mp.Len())

	l = mustGet(13).List()
	require.Equal(t, 0, l.Len())

	require.True(t, mustGet(14).IsVoid())

	runtime.GC()
	runtime.GC()
}

func TestList_Path(t *testing.T) {
	doc := automerge.New()

	err := doc.Path("y", 0).Set("a")
	require.NoError(t, err)
	err = doc.Path("y", 0).Set("b")
	require.NoError(t, err)
	b, err := automerge.As[string](doc.Path("y", 0).Get())
	require.NoError(t, err)
	require.Equal(t, "b", b)

	err = doc.Path("z").List().Append("a")
	require.NoError(t, err)
	err = doc.Path("z").List().Append("b")
	require.NoError(t, err)
	b, err = automerge.As[string](doc.Path("z", 1).Get())
	require.NoError(t, err)
	require.Equal(t, "b", b)

	b, err = automerge.As[string](doc.Path("z").List().Get(1))
	require.NoError(t, err)
	require.Equal(t, "b", b)

	require.Equal(t, 2, doc.Path("z").List().Len())

	i := doc.Path("y").List().Iter()
	for {
		i, v, valid := i.Next()
		if !valid {
			break
		}
		require.Equal(t, i, 0)
		require.Equal(t, v.Str(), "b")
	}

	require.NoError(t, i.Error())

	i = doc.Path("no").List().Iter()
	for {
		_, _, valid := i.Next()
		if !valid {
			break
		}
		t.Fatal("expected non-extant list to have no items")
	}
	require.NoError(t, i.Error())

	i = doc.Path("y", 0).List().Iter()
	for {
		_, _, valid := i.Next()
		if !valid {
			break
		}
		t.Fatal("expected non-list to have no items")
	}
	require.EqualError(t, i.Error(), "&automerge.Path{\"y\", 0}: tried to interate over non-list \"b\"")

	v, err := doc.Path("no", 3).Get()
	require.NoError(t, err)
	require.True(t, v.IsVoid())
}

func TestCounter(t *testing.T) {
	doc := automerge.New()

	c := automerge.NewCounter(10)
	require.Equal(t, "&automerge.Counter{10}", c.GoString())
	require.NoError(t, doc.RootMap().Set("x", c))
	require.NoError(t, c.Inc(-20))

	require.Equal(t, "&automerge.Counter{-10}", c.GoString())

	b := automerge.NewCounter(5)
	require.NoError(t, doc.Path("y", 0).Set(b))
	require.NoError(t, b.Inc(20))

	v, err := automerge.As[int8](doc.RootMap().Get("x"))
	require.Equal(t, int8(-10), v)
	require.NoError(t, err)

	v, err = automerge.As[int8](doc.Path("y", 0).Get())
	require.Equal(t, int8(25), v)
	require.NoError(t, err)

	_, err = automerge.NewCounter(10).Get()
	require.EqualError(t, err, "automerge.Counter: tried to read from detached counter")

	err = automerge.NewCounter(0).Inc(10)
	require.EqualError(t, err, "automerge.Counter: tried to write to detached counter")

	require.NoError(t, doc.RootMap().Set("x", true))
	_, err = c.Get()
	require.EqualError(t, err, "automerge.Counter: tried to read non-counter true")
}

func TestCounter_Path(t *testing.T) {
	doc := automerge.New()

	require.NoError(t, doc.Path("c").Counter().Inc(10))
	require.NoError(t, doc.Path("c").Counter().Inc(10))
	v, err := automerge.As[int64](doc.Path("c").Get())
	require.NoError(t, err)
	require.Equal(t, int64(20), v)

	v, err = doc.Path("b").Counter().Get()
	require.NoError(t, err)
	require.Equal(t, int64(0), v)

	require.NoError(t, doc.Path("list").Set([]any{}))
	err = doc.Path("list").Counter().Inc(10)
	require.EqualError(t, err, "&automerge.Path{\"list\"}: tried to increment non-counter &automerge.List{}")
}

func TestText(t *testing.T) {
	doc := automerge.New()

	txt := automerge.NewText("hello world")
	require.NoError(t, doc.RootMap().Set("x", txt))

	doc2, err := doc.Fork(nil)
	require.NoError(t, err)

	txt2, err := automerge.As[*automerge.Text](doc2.RootMap().Get("x"))
	require.NoError(t, err)

	require.NoError(t, txt.Append("!"))
	require.NoError(t, txt2.Insert(6, "cool "))

	_, err = doc.Merge(doc2)
	require.NoError(t, err)

	v, err := automerge.As[string](doc.RootMap().Get("x"))
	require.NoError(t, err)
	require.Equal(t, "hello cool world!", v)

	err = txt.Splice(100, 150, "test")
	require.EqualError(t, err, "automerge.Text: failed to write: Invalid pos 100")

	err = txt.Delete(6, 5)
	require.NoError(t, err)
	v, err = automerge.As[string](doc.RootMap().Get("x"))
	require.NoError(t, err)
	require.Equal(t, "hello world!", v)
	require.Equal(t, 12, txt.Len())

	require.Equal(t, "&automerge.Text{\"hello world!\"}", txt.GoString())
}

func TestText_Path(t *testing.T) {
	doc := automerge.New()

	require.NoError(t, doc.Path("text").Text().Set("hello world"))
	require.NoError(t, doc.Path("text").Text().Append("!"))
	require.Equal(t, 12, doc.Path("text").Text().Len())

	s, err := doc.Path("text").Text().Get()
	require.NoError(t, err)
	require.Equal(t, "hello world!", s)

	v, err := doc.Path("empty").Text().Get()
	require.NoError(t, err)
	require.Equal(t, "", v)

	s, err = automerge.As[string](doc.Path("text").Get())
	require.NoError(t, err)
	require.Equal(t, "hello world!", s)

	require.NoError(t, doc.Path("int").Set(10))
	err = doc.Path("int").Text().Append("!")
	require.EqualError(t, err, "&automerge.Path{\"int\"}: tried to edit non-text 10")
}

type Sandwich struct {
	Bread   string
	Filling []string
}

func TestChanges(t *testing.T) {
	doc := automerge.New()

	heads, err := doc.Heads()
	require.NoError(t, err)

	require.NoError(t, doc.Path("test").Set(&Sandwich{"rye", []string{"pastrami", "mustard"}}))
	changes, err := doc.Changes(heads)
	require.NoError(t, err)

	doc2 := automerge.New()
	require.NoError(t, doc2.Path("wow").Set(&Sandwich{Bread: "dutch crunch", Filling: []string{"brie", "cranberry"}}))

	require.NoError(t, doc2.Apply(changes))

	require.Equal(t, 2, doc2.RootMap().Len())
	v, err := automerge.As[map[string]*Sandwich](doc2.Root())
	require.NoError(t, err)
	require.Equal(t, "rye", v["test"].Bread)
	require.Equal(t, []string{"brie", "cranberry"}, v["wow"].Filling)

	bytes := changes.Save()
	require.NoError(t, err)
	changes, err = automerge.LoadChanges(bytes)
	require.NoError(t, err)

	doc3 := automerge.New()
	require.NoError(t, doc3.Apply(changes))

	require.Equal(t, 1, doc3.RootMap().Len())
	keys, err := doc3.RootMap().Keys()
	require.NoError(t, err)
	require.Equal(t, []string{"test"}, keys)

	doc4 := automerge.New()
	changes, err = doc2.Changes(nil)
	require.NoError(t, err)
	require.NoError(t, doc4.Apply(changes))

	v4, err := automerge.As[map[string]*Sandwich](doc2.Root())
	require.NoError(t, err)
	require.Equal(t, v, v4)
}

func TestIncremental(t *testing.T) {
	doc := automerge.New()

	b, err := doc.Save()
	require.NoError(t, err)

	doc2, err := automerge.Load(b)
	require.NoError(t, err)

	require.NoError(t, doc.Path("wow").Set(automerge.NewCounter(10)))

	b, err = doc.SaveIncremental()
	require.NoError(t, err)
	require.NoError(t, doc2.LoadIncremental(b))

	require.NoError(t, doc.Path("wow").Counter().Inc(10))

	b, err = doc.SaveIncremental()
	require.NoError(t, err)
	require.NoError(t, doc2.LoadIncremental(b))

	v, err := doc2.Path("wow").Counter().Get()
	require.NoError(t, err)
	require.Equal(t, int64(20), v)

	require.NoError(t, doc.Path("wow").Counter().Inc(10))
	require.NoError(t, doc2.Path("wow").Counter().Inc(-5))

	b, err = doc.SaveIncremental()
	require.NoError(t, err)
	require.NoError(t, doc2.LoadIncremental(b))

	v, err = doc2.Path("wow").Counter().Get()
	require.NoError(t, err)
	require.Equal(t, int64(25), v)
}

func TestSyncState(t *testing.T) {
	sDoc := automerge.New()
	sState := automerge.NewSyncState(sDoc)

	cDoc := automerge.New()
	cState := automerge.NewSyncState(cDoc)

	require.NoError(t, sDoc.Path("s").Counter().Inc(10))
	require.NoError(t, cDoc.Path("c").Counter().Inc(10))

	resync := func() {
		var valid1, valid2 bool
		var m []byte
		var err error

		for {
			m, valid1, err = cState.GenerateMessage()
			require.NoError(t, err)

			if !valid1 {
				break
			}
			require.NoError(t, sState.ReceiveMessage(m))

			m, valid2, err = sState.GenerateMessage()
			require.NoError(t, err)
			if !valid2 {
				break
			}

			require.NoError(t, cState.ReceiveMessage(m))
		}
	}

	resync()
	cV, err := automerge.As[map[string]int](cDoc.Root())
	require.NoError(t, err)
	sV, err := automerge.As[map[string]int](sDoc.Root())
	require.NoError(t, err)
	require.Equal(t, cV, sV)
	require.Equal(t, map[string]int{"s": 10, "c": 10}, cV)

	require.NoError(t, sDoc.Path("s").Counter().Inc(5))
	require.NoError(t, cDoc.Path("c").Counter().Inc(5))

	resync()
	resync()
	cV, err = automerge.As[map[string]int](cDoc.Root())
	require.NoError(t, err)
	sV, err = automerge.As[map[string]int](sDoc.Root())
	require.NoError(t, err)
	require.Equal(t, cV, sV)
	require.Equal(t, map[string]int{"s": 15, "c": 15}, cV)

	cBytes, err := cState.Save()
	require.NoError(t, err)
	sBytes, err := sState.Save()
	require.NoError(t, err)
	cState, err = automerge.LoadSyncState(cDoc, cBytes)
	require.NoError(t, err)
	sState, err = automerge.LoadSyncState(sDoc, sBytes)
	require.NoError(t, err)
	resync()

	require.NoError(t, sDoc.Path("s").Counter().Inc(5))
	require.NoError(t, cDoc.Path("c").Counter().Inc(5))

	resync()
	cV, err = automerge.As[map[string]int](cDoc.Root())
	require.NoError(t, err)
	sV, err = automerge.As[map[string]int](sDoc.Root())
	require.NoError(t, err)
	require.Equal(t, cV, sV)
	require.Equal(t, map[string]int{"s": 20, "c": 20}, cV)
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
