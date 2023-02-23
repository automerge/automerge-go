Automerge debug is a tool to help inspect automerge files (developed to
make it easier to create test cases for automerge).

It contains a tolerant parser for (a majority of) the file format
that can explain how each byte is interpreted, and highlight many potential errors.

Use at your own risk, and expect to have to edit the code and refer
to the [Binary Document
Format](https://alexjg.github.io/automerge-storage-docs/) to get good results.

To install: `go install github.com/automerge/automerge-go/cmd/automerge-debug@latest`

```
usage: automerge-debug [FLAGS] <file>?
automerge-debug outputs a commented byte stream of the encoded automerge chunk.

If no file is provided, the input is read from stdin.

If you have the bytes of a chunk's contents, but not the header, you can pass
--prefix=doc or --prefix=change to prefix your bytes with the correct header.

The output is designed to be valid go syntax containing every input byte, you can
edit this directly and then pass it through [bytes](github.com/ConradIrwin/bytes)
to convert back into binary.
  -compress
    	compress chunk
  -decompress
    	decompress chunk
  -fix-checksum
    	fix checksum
  -prefix string
    	prefix the bytes with a valid header (doc or change)
  -raw
    	output in binary
```

For example, when run on the [counter_value_is_overlong.automerge](https://github.com/automerge/automerge/blob/2cd7427f35e3b9b4a6b4d22d21dd083872015b57/rust/automerge/tests/fixtures/counter_value_is_overlong.automerge) test case, you get:

```
[]byte{133, 111, 74, 131, // magic bytes (valid)
111, 205, 220, 125, // checksum (valid)
1, // type = CHANGE CHUNK
53, // length = 53
    0, // number of heads = 0
    16, // id length = 16
    139, 6, 109, 163, 36, 47, 70, 96, 161, 100, 111, 136, 169, 17, 241, 148, // actor ID = 8b066da3242f4660a1646f88a911f194
    1, // sequence number = 1
    1, // start op = 1
    180, 182, 210, 208, 226, 48, // time = 2023-02-06 21:13:58.964 -0700 MST
    0, // message len = 0
    // commit message = ""
    0, // number of actor ids = 0
    6, // number of operation columns = 6
        21, 3, // column (spec = 21, id = 1, type = 5, deflate = false, length = 3)
        52, 1, // column (spec = 52, id = 3, type = 4, deflate = false, length = 1)
        66, 2, // column (spec = 66, id = 4, type = 2, deflate = false, length = 2)
        86, 2, // column (spec = 86, id = 5, type = 6, deflate = false, length = 2)
        87, 2, // column (spec = 87, id = 5, type = 7, deflate = false, length = 2)
        112, 2, // column (spec = 112, id = 7, type = 0, deflate = false, length = 2)
    // key string column
        127, // 1 literal values
            1, 97, // "a"
    // insert column
        1, // false repeated 1 times
    // action column
        127, // 1 literal values
            1, // 1
    // value meta column
        127, // 1 literal values
            40, // 40 (length = 2, type = 8)
    // value column
        208, 127, // counter -48(uleb error: overly long LEB)
    // predecessor group column
        127, // 1 literal values
            0, // 0
}
```
