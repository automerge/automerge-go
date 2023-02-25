package main

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// from: https://github.com/aviate-labs/leb128/blob/v0.1.0/leb.go#L13
func leb128(n uint64) []byte {
	leb := make([]byte, 0)
	if n == 0 {
		return []byte{0}
	}
	for n != 0x00 {
		b := byte(n & 0x7F)
		n >>= 7
		if n != 0x00 {
			b |= 0x80
		}
		leb = append(leb, b)
	}
	return leb
}

func readULEB(b []byte) (uint64, []byte, error) {
	var n uint64
	var i int
	for {
		if len(b) == 0 {
			return 0, nil, fmt.Errorf("failed to find end of LEB")
		}

		c := b[0]
		b = b[1:]
		n |= uint64(c&0x7f) << (i * 7)
		i++
		if c&0x80 == 0 {
			if c == 0 && i > 1 {
				return n, b, fmt.Errorf("overly long LEB")
			}
			if i > 10 {
				return n, b, fmt.Errorf("LEB > 64-bit")
			}
			if i == 10 && c&0x7e > 0 {
				return n, b, fmt.Errorf("LEB > 64-bit")
			}
			return n, b, nil
		}
	}
}

func readSLEB(b []byte) (int64, []byte, error) {
	var n uint64
	var i int
	var prev byte
	for {
		if len(b) == 0 {
			return 0, nil, fmt.Errorf("failed to find end of LEB")
		}

		c := b[0]
		b = b[1:]
		n |= uint64(c&0x7f) << (i * 7)
		i++
		if c&0x80 == 0 {
			if c&0x40 > 0 {
				n |= (math.MaxUint64 << (i * 7))
			}
			if i > 1 && (prev&0x40 == 0 && c == 0 || prev&0x40 > 0 && c == 0x7f) {
				return int64(n), b, fmt.Errorf("overly long LEB")
			}
			if i > 10 {
				return int64(n), b, fmt.Errorf("LEB > 64-bit")
			}
			if i == 10 && c&0x7f != 0 && c&0x7f != 0x7f {
				return int64(n), b, fmt.Errorf("LEB > 64-bit")
			}
			return int64(n), b, nil
		}
		prev = c
	}
}

var magicBytes = [4]byte{133, 111, 74, 131}

func addHeader(b []byte, kind byte) []byte {
	encodedLength := leb128(uint64(len(b)))
	payload := append(append([]byte{kind}, encodedLength...), b...)
	hash := sha256.Sum256(payload)
	return append(append(magicBytes[:], hash[:4]...), payload...)
}

func main() {
	prefix := ""
	raw := false
	compress := false
	decompress := false
	checksum := false

	flag.Usage = func() {
		fmt.Print(`usage: automerge-debug [--prefix=<doc|change>] [--raw] <file>?
automerge-debug outputs a commented byte stream of the encoded automerge chunk.

If no file is provided, the input is read from stdin.

If you have the bytes of a chunk's contents, but not the header, you can pass
--prefix=doc or --prefix=change to prefix your bytes with the correct header.

The output is designed to be valid go syntax containing every input byte, you can
edit this directly and then pass it through [bytes](github.com/ConradIrwin/bytes)
to convert back into binary.
`)

		flag.PrintDefaults()
	}
	flag.StringVar(&prefix, "prefix", "", "prefix the bytes with a valid header (doc or change)")
	flag.BoolVar(&raw, "raw", false, "output in binary")
	flag.BoolVar(&compress, "compress", false, "compress chunk")
	flag.BoolVar(&decompress, "decompress", false, "decompress chunk")
	flag.BoolVar(&checksum, "fix-checksum", false, "fix checksum")
	flag.Parse()

	var file []byte
	var err error
	if len(flag.Args()) == 0 {
		file, err = io.ReadAll(os.Stdin)
	} else {
		file, err = os.ReadFile(flag.Args()[0])
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if strings.HasPrefix(prefix, "doc") {
		file = addHeader(file, '\x00')
	} else if strings.HasPrefix(prefix, "ch") {
		file = addHeader(file, '\x01')
	} else if checksum {
		file = fixChecksum(file)
	}

	if compress {
		file = compressChunk(file)
	} else if decompress {
		file = decompressChunk(file)
	}

	if raw {
		os.Stdout.Write(file)
	} else {
		prettyPrintBytes(file)
	}
}

func inputErr(s string, args ...any) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
	os.Exit(1)
}

func validateChunk(b []byte, expected byte) []byte {
	if len(b) < 10 {
		inputErr("not a valid chunk (too short)")
	}
	if b[0] != 133 || b[1] != 111 || b[2] != 74 || b[3] != 131 {
		inputErr("not a valid chunk (invalid magic bytes)")
	}

	t := b[8]
	if t != expected && expected != 255 {
		if t == 1 && expected == 2 {
			inputErr("already decompressed")
		} else if t == 2 && expected == 1 {
			inputErr("already compressed")
		} else {
			inputErr("wrong chunk type (type = %v, expected = %v)", t, expected)
		}
	}

	i, chunk, err := readULEB(b[9:])
	if err != nil {
		inputErr("not a valid change chunk (incorrectly encoded length)")
	}
	if len(chunk) != int(i) {
		inputErr("not a valid change chunk (trailing data)")
	}

	return chunk
}

func compressChunk(b []byte) []byte {
	chunk := validateChunk(b, 1)

	out := &bytes.Buffer{}
	w, err := flate.NewWriter(out, 9)
	if err != nil {
		inputErr("failed to compress: %v", err)
	}
	_, err = w.Write(chunk)
	if err != nil {
		inputErr("failed to compress: %v", err)
	}
	if err := w.Close(); err != nil {
		inputErr("failed to compress: %v", err)
	}

	content := out.Bytes()

	ret := &bytes.Buffer{}

	ret.Write(b[0:8])
	ret.Write([]byte{2})
	ret.Write(leb128(uint64(len(content))))
	ret.Write(content)
	return ret.Bytes()
}

func decompressChunk(b []byte) []byte {
	chunk := validateChunk(b, 2)

	r := flate.NewReader(bytes.NewReader(chunk))

	out, err := io.ReadAll(r)
	if err != nil {
		inputErr("failed to decompress: %v", err)
	}

	ret := &bytes.Buffer{}
	ret.Write(b[0:8])
	ret.Write([]byte{1})
	ret.Write(leb128(uint64(len(out))))
	ret.Write(out)
	return ret.Bytes()
}

func fixChecksum(b []byte) []byte {
	validateChunk(b, 255)
	compressed := b[8] == 2

	if compressed {
		b = decompressChunk(b)
	}

	if b[8] == 0 {
		// TODO: fix change hashes too? maybe add a mode to convert from doc to changes and back
		fmt.Fprintf(os.Stderr, "refusing to fix checksum of document chunk - your change hashes are probably wrong too")
		os.Exit(1)
	}

	hash := sha256.Sum256(b[8:])
	b[4] = hash[0]
	b[5] = hash[1]
	b[6] = hash[2]
	b[7] = hash[3]

	if compressed {
		b = compressChunk(b)
	}

	return b
}

func prettyPrintBytes(b []byte) {
	fmt.Print("[]byte{")

	for len(b) > 0 {
		b = prettyPrintChunk(b)
	}

	fmt.Print("}\n")
}

func prettyPrintChunk(b []byte) []byte {
	if len(b) < 10 {
		pBytes(b, "ERROR: too short (expected magic bytes, checksum, type and length)")
		return nil
	}
	if b[0] != 133 || b[1] != 111 || b[2] != 74 || b[3] != 131 {
		pBytes(b[0:4], "ERROR: invalid magic bytes")
	} else {
		pBytes(b[0:4], "magic bytes (valid)")
	}

	t := b[8]
	i, chunk, err := readULEB(b[9:])
	lebLength := len(b) - len(chunk) - 9

	chunkEnd := 8 + 1 + lebLength + int(i)

	if len(b) < chunkEnd {
		chunkEnd = len(b)
		err = fmt.Errorf("longer than remaining data: %v", len(b))
	}

	hash := sha256.Sum256(b[8:chunkEnd])

	if t == 2 {
		pBytes(b[4:8], "checksum (not validated)")
	} else if hash[0] != b[4] || hash[1] != b[5] || hash[2] != b[6] || hash[3] != b[7] {
		pBytes(b[4:8], "ERROR: checksum mis-match (should be"+formatBytes(hash[0:4])+")")
	} else {
		pBytes(b[4:8], "checksum (valid)")
	}

	if t == 0 {
		pBytes(b[8:9], "type = DOCUMENT CHUNK")
	} else if t == 1 {
		pBytes(b[8:9], "type = CHANGE CHUNK")
	} else if t == 2 {
		pBytes(b[8:9], " type = COMPRESSED CHUNK")
	} else {
		pBytes(b[8:9], "type = INVALID (should be 0, 1 or 2)")
	}

	if err != nil {
		pBytes(b[9:9+lebLength], "length = "+fmt.Sprint(i)+" (error: "+err.Error()+")")
	}
	pBytes(b[9:9+lebLength], "length = "+fmt.Sprint(i))

	if t == 0 {
		prettyPrintDoc(chunk)
	} else if t == 1 {
		prettyPrintChange(chunk)
	} else if t == 2 {
		pBytes(chunk, "DEFLATE stream")
	} else {
		pBytes(chunk, "chunk data")
	}

	if 9+lebLength+int(i) >= len(b) {
		return nil
	}

	return b[9+lebLength+int(i):]
}

func prettyPrintDoc(b []byte) {
	defer indent()()
	_, b = printActorIDs(b)
	hs, b := printHeads(b)
	cc, b := printColumnMeta(b, "change")
	oc, b := printColumnMeta(b, "operation")

	for _, c := range cc {
		if len(b) < c.Length {
			pBytes(b, "ERROR: not enough bytes to read column length = %d", c.Length)
			return
		}
		printChangeColumn(b[:c.Length], c)
		b = b[c.Length:]
	}

	for _, c := range oc {
		if len(b) < c.Length {
			pBytes(b, "ERROR: not enough bytes to read column length = %d", c.Length)
			return
		}
		printOperationColumn(b[:c.Length], c)
		b = b[c.Length:]
	}

	b = printHeadIndexes(b, hs)

	if len(b) > 0 {
		pBytes(b, "remaining data in chunk")
	}
}

func prettyPrintChange(b []byte) {
	defer indent()()
	_, b = printHeads(b)
	_, b = printActorID(b)

	errS := func(err error) string {
		if err == nil {
			return ""
		} else {
			return "(error: " + err.Error() + ")"
		}
	}

	seq, rest, err := readULEB(b)
	pBytes(b[0:len(b)-len(rest)], "sequence number = %v %s", seq, errS(err))
	b = rest
	sop, rest, err := readULEB(b)
	pBytes(b[0:len(b)-len(rest)], "start op = %v %s", sop, errS(err))
	b = rest
	ms, rest, err := readSLEB(b)
	if ms == 0 {
		pBytes(b[0:len(b)-len(rest)], "time = %v %s", 0, errS(err))
	} else {
		pBytes(b[0:len(b)-len(rest)], "time = %v %s", time.UnixMilli(ms), errS(err))
	}
	b = rest

	mlen, rest, err := readULEB(b)
	if mlen > uint64(len(rest)) {
		pBytes(b[0:len(b)-len(rest)], "message len = %v (error: too long) %s", mlen, errS(err))
		pBytes(rest, "commit message (truncated)")
		return
	}
	pBytes(b[0:len(b)-len(rest)], "message len = %v %s", mlen, errS(err))
	pBytes(rest[:mlen], "commit message = %#v", string(rest[:mlen]))
	b = rest[mlen:]
	_, b = printActorIDs(b)

	oc, b := printColumnMeta(b, "operation")
	for _, c := range oc {
		if len(b) < c.Length {
			pBytes(b, "ERROR: not enough bytes to read column length = %d", c.Length)
			return
		}
		printOperationColumn(b[:c.Length], c)
		b = b[c.Length:]
	}

	if len(b) > 0 {
		pBytes(b, "extra bytes")
	}
}

func printActorID(b []byte) ([]byte, []byte) {
	l, rest, err := readULEB(b)
	if err != nil {
		pBytes(b[0:len(b)-len(rest)], "id length = "+fmt.Sprint(l)+"(error: "+err.Error()+")")
	} else {
		pBytes(b[0:len(b)-len(rest)], "id length = "+fmt.Sprint(l))
	}

	if l > math.MaxInt || int(l) > len(rest) {
		pBytes(rest, "ERROR: actor id length greater than remains in document")
		return nil, nil
	}
	pBytes(rest[0:l], "actor ID = "+hex.EncodeToString(rest[0:l]))
	return rest[0:l], rest[l:]
}

func printActorIDs(b []byte) ([][]byte, []byte) {
	actorIDs := [][]byte{}

	n, rest, err := readULEB(b)
	if err != nil {
		pBytes(b[0:len(b)-len(rest)], "number of actor ids = "+fmt.Sprint(n)+"(error: "+err.Error()+")")
	} else {
		pBytes(b[0:len(b)-len(rest)], "number of actor ids = "+fmt.Sprint(n))
	}

	defer indent()()

	b = rest
	for i := uint64(0); i < n; i++ {
		actorID, rest := printActorID(b)
		if actorID != nil {
			actorIDs = append(actorIDs, actorID)
		}
		b = rest
	}
	return actorIDs, b
}

func printHeads(b []byte) ([][]byte, []byte) {
	heads := [][]byte{}
	n, rest, err := readULEB(b)
	if err != nil {
		pBytes(b[0:len(b)-len(rest)], "number of heads = "+fmt.Sprint(n)+"(error: "+err.Error()+")")
	} else {
		pBytes(b[0:len(b)-len(rest)], "number of heads = "+fmt.Sprint(n))
	}
	b = rest

	defer indent()()

	for n > 0 {
		if len(b) == 0 {
			pBytes(b, " ERROR: missing head")
			return heads, nil
		}
		if len(b) < 32 {
			pBytes(b, " ERROR: incomplete head")
			return heads, nil
		}

		heads = append(heads, b[0:32])
		pBytes(b[0:32], " head "+hex.EncodeToString(b[0:32]))
		b = b[32:]
		n--
	}

	return heads, b
}

func printHeadIndexes(b []byte, hs [][]byte) []byte {
	if len(b) == 0 {
		pBytes(nil, "head index omitted")
		return b
	}
	pBytes(nil, "head index")
	defer indent()()
	for _, h := range hs {
		n, rest, err := readULEB(b)
		if err != nil {
			pBytes(b[0:len(b)-len(rest)], "%s is change %d (error: %s)", hex.EncodeToString(h), n, err.Error())
		} else {
			pBytes(b[0:len(b)-len(rest)], "%s is change %d", hex.EncodeToString(h), n)
		}
		b = rest
	}
	return b
}

type Column struct {
	Length  int
	Spec    int32
	ID      int32
	Type    int8
	Deflate bool
}

func printColumnMeta(b []byte, kind string) ([]Column, []byte) {
	cols := []Column{}
	n, rest, err := readULEB(b)
	if err != nil {
		pBytes(b[0:len(b)-len(rest)], "number of "+kind+" columns = "+fmt.Sprint(n)+"(error: "+err.Error()+")")
	} else {
		pBytes(b[0:len(b)-len(rest)], "number of "+kind+" columns = "+fmt.Sprint(n))
	}
	b = rest

	defer indent()()

	for n > 0 {
		spec, rest, err := readULEB(b)
		l, rest, err2 := readULEB(rest)
		repr := b[0 : len(b)-len(rest)]
		b = rest

		col := Column{
			Spec:    int32(spec),
			ID:      int32(spec >> 4),
			Type:    int8(spec & 0x07),
			Deflate: spec&0x08 > 0,
			Length:  int(l),
		}
		errS := ""
		if err != nil {
			errS = "(error in spec: " + err.Error() + ") "
		} else if spec > math.MaxUint32 {
			errS = "(error: spec is too large) "
		}
		if err2 != nil {
			errS += "(error in length: " + err2.Error() + ") "
		} else if l > math.MaxInt {
			errS = "(error: length is too long)"
		}

		pBytes(repr, "column (spec = %d, id = %d, type = %d, deflate = %v, length = %d) %s", spec, col.ID, col.Type, col.Deflate, col.Length, errS)
		cols = append(cols, col)
		n--
	}

	return cols, b
}

var changeCols = map[int32]string{
	1:  "actor",
	3:  "sequence number",
	19: "maxOp",
	35: "time",
	53: "message",
	64: "dependencies group",
	67: "dependencies index",
	86: "extra metadata",
	87: "extra data",
}

var opCols = map[int32]string{
	1:   "object actor id",
	2:   "object counter",
	17:  "key actor id",
	19:  "key counter",
	21:  "key string",
	33:  "actor id",
	35:  "counter",
	52:  "insert",
	66:  "action",
	86:  "value meta",
	87:  "value",
	112: "predecessor group",
	113: "predecessor actor id",
	115: "predecessor counter",
	128: "successor group",
	129: "successor actor id",
	131: "successor counter",
}

func printChangeColumn(b []byte, c Column) {
	title := changeCols[c.Spec]
	if title == "" {
		title = fmt.Sprintf("Unknown (id=%d, type=%d)", c.ID, c.Type)
	}
	pBytes(nil, "%v column", title)
	defer indent()()
	printColumn(b, c)
}

func printOperationColumn(b []byte, c Column) {
	title := opCols[c.Spec]
	if title == "" {
		title = fmt.Sprintf("Unknown (id=%d, type=%d)", c.ID, c.Type)
	}
	pBytes(nil, "%v column", title)
	defer indent()()
	printColumn(b, c)
}

type Meta struct {
	Length int
	Type   int8
}

var metaColumn []Meta
var metaID int32

func printColumn(b []byte, c Column) {
	if c.Deflate {
		pBytes(b, "Compresed column")
		return
	}

	switch c.Type {
	// case 0:
	// grouped column!
	case 0, 1, 2:
		printUlebColumn(b)
	case 3:
		printDeltaColumn(b)
	case 4:
		printBoolColumn(b)
	case 5:
		printStringColumn(b)
	case 6:
		metaColumn = printValueMetaColumn(b)
		metaID = c.ID
	case 7:
		if metaID == c.ID {
			printValueColumn(b, metaColumn)
		} else {
			pBytes(b, "ERROR: value column with no metadata", c.Type)
		}

	default:
		pBytes(b, "ERROR: unknown column format %d", c.Type)
	}
}

func printUlebColumn(b []byte) {
	printRunLengthEncoded(b, func(b []byte, n int64) (string, []byte) {
		v, rest, err := readULEB(b)
		ret := fmt.Sprint(v)
		if err != nil {
			ret += " (value error: " + err.Error() + ")"
		}

		return ret, rest
	})
}

func printDeltaColumn(b []byte) {
	i := int64(0)

	for len(b) > 0 {
		n, rest, err := readSLEB(b)
		errS := ""
		if err != nil {
			errS = "(length error: " + err.Error() + ") "
		}

		if n == 0 {
			l, rest, err := readULEB(rest)
			if err != nil {
				errS += "(null error: " + err.Error() + ")"
			}

			pBytes(b[:len(b)-len(rest)], "null repeated %d times %s", l, errS)
			b = rest
			continue
		}

		if n > 0 {
			delta, rest, err := readSLEB(rest)
			if err != nil {
				errS += "(delta error: " + err.Error() + ")"
			}

			nextI := int64(i) + delta
			finalI := int64(i) + delta*n

			if finalI < 0 {
				errS += "(rle error: i < 0)"
			}

			pBytes(b[:len(b)-len(rest)], "%d,... %d (%d steps of %d) %s", nextI, finalI, n, delta, errS)
			b = rest
			i = int64(finalI)
			continue
		}

		pBytes(b[:len(b)-len(rest)], "%d literal deltas %s", 0-n, errS)
		undo := indent()

		b = rest
		for n < 0 {
			delta, rest, err := readSLEB(b)
			errS = ""
			if err != nil {
				errS = "(value error: " + err.Error() + ")"
			}
			i += delta

			if i < 0 {
				errS += "(rle error: i < 0)"
			}

			pBytes(b[:len(b)-len(rest)], "%+d = %d %s", delta, i, errS)

			b = rest
			n++
		}

		undo()

	}
}

func printBoolColumn(b []byte) {
	curr := false
	for len(b) > 0 {
		n, rest, err := readULEB(b)
		errS := ""
		if err != nil {
			errS = "(length error: " + err.Error() + ") "
		}
		pBytes(b[:len(b)-len(rest)], "%v repeated %v times %s", curr, n, errS)
		curr = !curr
		b = rest
	}
}

func printStringColumn(b []byte) {
	printRunLengthEncoded(b, func(b []byte, n int64) (string, []byte) {
		l, rest, err := readULEB(b)
		errS := ""
		if err != nil {
			errS += "(value error: " + err.Error() + ")"
		}
		sl := int(l)

		if uint64(len(rest)) < l {
			errS += "(string error: not enough bytes remaining)"
			sl = len(rest)
		}

		s := rest[:sl]
		if !utf8.Valid(s) {
			errS += "(utf8 error: invalid utf-8)"
			s = []byte(string([]rune(string(s))))
		}

		return fmt.Sprintf("%#v %s", string(s), errS), rest[sl:]
	})
}

func printValueMetaColumn(b []byte) []Meta {
	meta := []Meta{}
	printRunLengthEncoded(b, func(b []byte, n int64) (string, []byte) {
		spec, rest, err := readULEB(b)
		errS := ""
		if err != nil {
			errS = " (meta error: " + err.Error() + ")"
		}

		m := Meta{
			Length: int(spec >> 4),
			Type:   int8(spec & 0x0F),
		}
		for i := int64(0); i < n; i++ {
			meta = append(meta, m)
		}

		return fmt.Sprintf("%d (length = %d, type = %d) %s", spec, m.Length, m.Type, errS), rest
	})
	return meta
}

func printValueColumn(b []byte, meta []Meta) {
	for _, m := range meta {
		s := ""
		exp := m.Length
		switch m.Type {
		case 0:
			s = "null"
			exp = 0
		case 1:
			s = "false"
			exp = 0
		case 2:
			s = "true"
			exp = 0
		case 3:
			u, rest, err := readULEB(b[:m.Length])
			s = fmt.Sprintf("uint %d", u)
			if err != nil {
				s += "(uleb error: " + err.Error() + ")"
			}
			exp = m.Length - len(rest)
		case 4:
			i, rest, err := readSLEB(b[:m.Length])
			s = fmt.Sprintf("uint %d", i)
			if err != nil {
				s += "(uleb error: " + err.Error() + ")"
			}
			exp = m.Length - len(rest)
		case 5:
			exp = 8
			if len(b) < 8 {
				pBytes(b, "ERROR: not enough bytes for float")
				break
			}
			bits := binary.LittleEndian.Uint64(b[:8])
			float := math.Float64frombits(bits)
			s = fmt.Sprintf("float %v", float)

		case 6:
			s = fmt.Sprintf("string %#v", string((b[:m.Length])))
			if !utf8.Valid(b[:m.Length]) {
				s += " (utf-8 error: invalid utf-8)"
			}

		case 7:
			s = fmt.Sprintf("bytes %#v", b[:m.Length])

		case 8:
			i, rest, err := readSLEB(b[:m.Length])
			s = fmt.Sprintf("counter %d", i)
			if err != nil {
				s += "(uleb error: " + err.Error() + ")"
			}
			exp = m.Length - len(rest)

		case 9:
			i, rest, err := readSLEB(b[:m.Length])
			s = fmt.Sprintf("time %v", time.UnixMilli(i).UTC().Format(time.RFC3339))
			if err != nil {
				s += "(uleb error: " + err.Error() + ")"
			}
			exp = m.Length - len(rest)
		default:
			s = fmt.Sprintf("(error unknown value type = %d", m.Type)
		}

		if m.Length != exp {
			s += fmt.Sprintf(" (error: value length %d != %d from metadata)", exp, m.Length)
		}

		pBytes(b[:m.Length], s)
		b = b[m.Length:]
	}
}

func printRunLengthEncoded(b []byte, f func(b []byte, n int64) (string, []byte)) {
	for len(b) > 0 {
		n, rest, err := readSLEB(b)
		errS := ""
		if err != nil {
			errS = "(length error: " + err.Error() + ") "
		}

		if n == 0 {
			l, rest, err := readULEB(rest)
			if err != nil {
				errS += "(null error: " + err.Error() + ")"
			}

			pBytes(b[:len(b)-len(rest)], "null repeated %d times %s", l, errS)
			b = rest
			continue
		}

		if n > 0 {
			s, rest := f(rest, n)
			pBytes(b[:len(b)-len(rest)], "%s repeated %d times %s", s, n, errS)
			b = rest
			continue
		}

		pBytes(b[:len(b)-len(rest)], "%d literal values %s", 0-n, errS)
		undo := indent()

		b = rest
		for n < 0 {
			s, rest := f(b, 1)
			pBytes(b[:len(b)-len(rest)], s)
			b = rest
			n++
		}

		undo()
	}
}

func formatBytes(b []byte) string {
	s := "["
	for i, c := range b {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%d", c)
	}
	return s + "]"
}

var prefix = ""

func indent() func() {
	oldPrefix := prefix
	prefix += "    "
	return func() { prefix = oldPrefix }
}

func pBytes(b []byte, desc string, args ...any) {
	fmt.Print(prefix)
	for _, c := range b {
		fmt.Printf("%d, ", c)
	}
	fmt.Printf("// "+desc+"\n", args...)
}
