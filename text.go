package automerge

// #include "automerge.h"
import "C"
import "fmt"

// Text is a mutable unicode string that can be edited collaboratively.
//
// Note that automerge considers text to be a sequence of unicode codepoints
// while most go code treats strings as a sequence of bytes (that are hopefully valid utf8).
// In go programs unicode codepoints are stored as integers of type [rune], and the [unicode/utf8] package
// provides some helpers for common operations.
//
// When editing Text you must pass positions and counts in terms of codepoints not bytes.
// For example if you wanted to replace the first instance of "🙃" you could do something like this:
//
//	s, _ := text.Get() => "😀🙃"
//	byteIndex := strings.Index(s, "🙃") => 4
//	runeIndex := utf8.RuneCountInString(s[:byteIndex]) => 1
//	text.Splice(runeIndex, 1, "🧟") => "🙃🧟"
//
// Although it is possible to represent invalid utf8 in a go string, automerge will error if you
// try to write invalid utf8 into a document.
//
// If you are new to unicode it's worth pointing out that the number of codepoints does not
// necessarily correspond to the number of rendered glyphs (for example Text("👍🏼").Len() == 2).
// For more information consult the Unicode Consortium's [FAQ].
//
// [FAQ]: https://www.unicode.org/faq/char_combmark.html
// [rune]: https://pkg.go.dev/builtin#rune
type Text struct {
	doc   *Doc
	objID *objID
	path  *Path

	val string
}

func (t *Text) lock() (*C.AMdoc, *C.AMobjId, func()) {
	cDoc, unlock := t.doc.lock()
	return cDoc, t.objID.cObjID, unlock
}

// NewText returns a detached Text with the given starting value.
// Before you can read or write it you must write it to the document.
func NewText(s string) *Text {
	return &Text{val: s}
}

// Len returns number of unicode codepoints in the text, this
// may be less than the number of utf-8 bytes.
// For example Text("😀😀").Len() == 2, while len("😀😀") == 8.
func (t *Text) Len() int {
	if t.doc == nil {
		return 0
	}
	if t.path != nil {
		v, err := t.path.Get()
		if err != nil || v.Kind() != KindText {
			return 0
		}
		return v.Text().Len()
	}

	cDoc, cObj, unlock := t.lock()
	defer unlock()
	return int(C.AMobjSize(cDoc, cObj, nil))
}

// Get returns the current value as a string
func (t *Text) Get() (string, error) {
	if t.doc == nil {
		return "", fmt.Errorf("automerge.Text: tried to read detached text")
	}
	if t.path != nil {
		v, err := t.path.Get()
		if err != nil {
			return "", err
		}
		switch v.Kind() {
		case KindVoid:
			return "", nil
		case KindText:
			return v.Text().Get()
		default:
			return "", fmt.Errorf("automerge.Text: tried to read non-text value %#v", v.val)
		}
	}

	cDoc, cObj, unlock := t.lock()
	defer unlock()

	s, err := wrap(C.AMtext(cDoc, cObj, nil)).item()
	if err != nil {
		return "", err
	}
	return s.str(), nil
}

// Set overwrites the entire string with a new value,
// prefer to use Insert/Delete/Append/Splice as appropriate
// to preserves collaborators changes.
func (t *Text) Set(s string) error {
	return t.splice(0, C.PTRDIFF_MAX, s)
}

// Insert adds a substr at position pos in the Text
func (t *Text) Insert(pos int, s string) error {
	return t.splice(C.size_t(pos), 0, s)
}

// Delete deletes del runes from position pos
func (t *Text) Delete(pos int, del int) error {
	return t.splice(C.size_t(pos), C.ptrdiff_t(del), "")
}

// Append adds substr s at the end of the string
func (t *Text) Append(s string) error {
	return t.splice(C.SIZE_MAX, 0, s)
}

// Splice deletes del runes at position pos, and inserts
// substr s in their place.
func (t *Text) Splice(pos int, del int, s string) error {
	return t.splice(C.size_t(pos), C.ptrdiff_t(del), s)
}

func (t *Text) splice(pos C.size_t, del C.ptrdiff_t, s string) error {
	if t.doc == nil {
		return fmt.Errorf("automerge.Text: tried to write to detached text")
	}
	if t.path != nil {
		t2, err := t.path.ensureText()
		if err != nil {
			return err
		}
		t.objID = t2.objID
		t.path = nil
	}

	cStr, free := toByteSpanStr(s)
	defer free()
	cDoc, cObj, unlock := t.lock()
	defer unlock()

	err := wrap(C.AMspliceText(cDoc, cObj, pos, del, cStr)).void()
	if err != nil {
		return fmt.Errorf("automerge.Text: failed to write: %w", err)
	}
	return nil
}

// GoString returns a representation suitable for debugging.
func (t *Text) GoString() string {
	if t.doc == nil {
		return fmt.Sprintf("&automerge.Text{%#v}", t.val)
	}
	v, err := t.Get()
	if err != nil {
		return "&automerge.Text{<error>}"
	}
	return fmt.Sprintf("&automerge.Text{%#v}", v)
}

// String returns a representation suitable for debugging.
func (t *Text) String() string {
	return t.GoString()
}
