package rwkvtkn

import (
	"testing"
)

func intSliceEquals(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// TestSimpleRoundtrip tests the creation of a tokenizer with the
// default vocabulary and round-tripping a unicode string.
func TestSimpleRoundtrip(t *testing.T) {
	tkn := NewWorldTokenizer()

	s := "Hello, world! こんにちは、世界！"
	i := []int{33155, 45, 40213, 34, 33, 10115, 10165, 10136, 10127, 10139, 10079, 10267, 14610, 19126}

	x, err := tkn.EncodeString(s)
	if !intSliceEquals(x, i) || err != nil {
		t.Fatalf(`EncodeString(%q) = %v, %v, want equal to %v`, s, x, err, i)
	}

	y, err := tkn.DecodeToString(x)
	if y != s || err != nil {
		t.Fatalf(`DecodeToString(%v) = %q, %v, want equal to %q`, x, y, err, s)
	}
}
