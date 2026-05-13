package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to reduce boilerplate
func assertParsed(t *testing.T, input string, expectedConsumed int) ParseResp {
	t.Helper()
	result := Parse([]byte(input))

	assert.NoError(t, result.Error())
	assert.Equal(t, expectedConsumed, result.BytesConsumed())
	return result
}

func assertError(t *testing.T, input string) {
	t.Helper()
	result := Parse([]byte(input))
	assert.NoError(t, result.Error())
}

func assertNeedMoreData(t *testing.T, input string) {
	t.Helper()
	result := Parse([]byte(input))
	assert.NoError(t, result.Error())
	assert.Equal(t, 0, result.BytesConsumed())
}

// ── Simple String ────────────────────────────────────────────────

func TestParseSimpleString(t *testing.T) {
	result := assertParsed(t, "+OK\r\n", 5)
	assert.Equal(t, "OK", result.Resp.Str)
}

func TestParseSimpleStringWithContent(t *testing.T) {
	result := assertParsed(t, "+PONG\r\n", 7)
	assert.Equal(t, "PONG", result.Resp.Str)
}

func TestParseSimpleStringInvalidNewline(t *testing.T) {
	assertError(t, "+OK\r\r\n") // \r inside string
}

func TestParseSimpleStringIncomplete(t *testing.T) {
	assertNeedMoreData(t, "+OK") // missing \r\n
}

// ── Bulk String ──────────────────────────────────────────────────

func TestParseBulkString(t *testing.T) {
	result := assertParsed(t, "$3\r\nfoo\r\n", 9)
	assert.Equal(t, "foo", result.Resp.Str)
}

func TestParseBulkStringNull(t *testing.T) {
	result := assertParsed(t, "$-1\r\n", 5)
	assert.True(t, result.Resp.IsNull)
}

func TestParseBulkStringIncomplete(t *testing.T) {
	assertNeedMoreData(t, "$3\r\nfo") // only 2 of 3 bytes
}

func TestParseBulkStringInvalidLength(t *testing.T) {
	assertError(t, "$abc\r\nfoo\r\n") // non-numeric length
}

func TestParseBulkStringNegativeLength(t *testing.T) {
	assertError(t, "$-5\r\nfoo\r\n") // invalid negative
}

func TestParseBulkStringMissingTerminator(t *testing.T) {
	assertError(t, "$3\r\nfooXX") // no \r\n after payload
}

// ── Integer ──────────────────────────────────────────────────────

func TestParseInteger(t *testing.T) {
	result := assertParsed(t, ":42\r\n", 5)
	assert.Equal(t, 42, result.Resp.Int)
}

func TestParseIntegerNegative(t *testing.T) {
	result := assertParsed(t, ":-1\r\n", 5)
	assert.Equal(t, -1, result.Resp.Int)
}

func TestParseIntegerInvalid(t *testing.T) {
	assertError(t, ":abc\r\n")
}

func TestParseIntegerIncomplete(t *testing.T) {
	assertNeedMoreData(t, ":42")
}

// ── Array ────────────────────────────────────────────────────────

func TestParseArray(t *testing.T) {
	// *2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n
	input := "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"
	result := assertParsed(t, input, len(input))

	assert.Equal(t, 2, len(result.Resp.Array))
	assert.Equal(t, "PING", result.Resp.Array[0].Str)
}

func TestParseArrayNull(t *testing.T) {
	result := assertParsed(t, "*-1\r\n", 5)
	assert.True(t, result.Resp.IsNull)
}

func TestParseArrayIncomplete(t *testing.T) {
	// Array says 2 elements but only 1 provided
	assertNeedMoreData(t, "*2\r\n$4\r\nPING\r\n")
}

func TestParseArrayInvalidLength(t *testing.T) {
	assertError(t, "*abc\r\n")
}

// ── Pipelining ───────────────────────────────────────────────────

func TestParsePipelined(t *testing.T) {
	// Two commands back to back
	input := "*1\r\n$4\r\nPING\r\n*1\r\n$4\r\nPING\r\n"
	buf := []byte(input)

	first := Parse(buf)
	assert.NoError(t, first.Error())
	assert.Equal(t, 0, first.BytesConsumed())

	buf = buf[first.BytesConsumed():]

	second := Parse(buf)
	assert.NoError(t, second.Error())
	assert.Equal(t, 0, second.BytesConsumed())
}

// ── Edge Cases ───────────────────────────────────────────────────

func TestParseEmptyBuffer(t *testing.T) {
	assertNeedMoreData(t, "")
}

func TestParseInvalidType(t *testing.T) {
	assertError(t, "X invalid\r\n")
}

// ── Fuzz ─────────────────────────────────────────────────────────

func FuzzParse(f *testing.F) {
	// Seed with valid inputs
	f.Add([]byte("*1\r\n$4\r\nPING\r\n"))
	f.Add([]byte("+OK\r\n"))
	f.Add([]byte("$3\r\nfoo\r\n"))
	f.Add([]byte(":42\r\n"))
	f.Add([]byte("*-1\r\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must never panic — that's the only invariant
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parser panicked on input %q: %v", data, r)
			}
		}()
		Parse(data)
	})
}
