package resp

import (
	"testing"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func TestEncodeSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "OK",
			input:    common.RespValue{Type: enums.SimpleStringRespType, Str: "OK"},
			expected: "+OK\r\n",
		},
		{
			name:     "PONG",
			input:    common.RespValue{Type: enums.SimpleStringRespType, Str: "PONG"},
			expected: "+PONG\r\n",
		},
		{
			name:     "empty string",
			input:    common.RespValue{Type: enums.SimpleStringRespType, Str: ""},
			expected: "+\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			if result != tt.expected {
				t.Fatalf("expected %q got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "protocol error",
			input:    common.RespValue{Type: enums.ErrorRespType, Str: "ERR Protocol error"},
			expected: "-ERR Protocol error\r\n",
		},
		{
			name:     "unknown command",
			input:    common.RespValue{Type: enums.ErrorRespType, Str: "ERR unknown command 'FOO'"},
			expected: "-ERR unknown command 'FOO'\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			if result != tt.expected {
				t.Fatalf("expected %q got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "positive",
			input:    common.RespValue{Type: enums.IntRespType, Int: 42},
			expected: ":42\r\n",
		},
		{
			name:     "zero",
			input:    common.RespValue{Type: enums.IntRespType, Int: 0},
			expected: ":0\r\n",
		},
		{
			name:     "negative",
			input:    common.RespValue{Type: enums.IntRespType, Int: -1},
			expected: ":-1\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			if result != tt.expected {
				t.Fatalf("expected %q got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "simple value",
			input:    common.RespValue{Type: enums.BulkStringRespType, Str: "foo"},
			expected: "$3\r\nfoo\r\n",
		},
		{
			name:     "empty string",
			input:    common.RespValue{Type: enums.BulkStringRespType, Str: ""},
			expected: "$0\r\n\r\n",
		},
		{
			name:     "longer value",
			input:    common.RespValue{Type: enums.BulkStringRespType, Str: "hello world"},
			expected: "$11\r\nhello world\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			if result != tt.expected {
				t.Fatalf("expected %q got %q", tt.expected, result)
			}
		})
	}
}

// Round trip — parse what encoder produces
func TestEncoderRoundTrip(t *testing.T) {
	values := []common.RespValue{
		{Type: enums.SimpleStringRespType, Str: "OK"},
		{Type: enums.IntRespType, Int: 42},
		{Type: enums.BulkStringRespType, Str: "hello"},
		{Type: enums.ErrorRespType, Str: "ERR something"},
	}

	for _, v := range values {
		encoded := Encoder(v)
		parsed := Parse(encoded)
		if parsed.Error() != nil {
			t.Fatalf("round trip failed for %+v: %v", v, parsed.Error())
		}
		if parsed.BytesConsumed() != len(encoded) {
			t.Fatalf("round trip consumed %d bytes, expected %d", parsed.BytesConsumed(), len(encoded))
		}
	}
}

// in encoder_test.go
func TestEncodeBulkStringNull(t *testing.T) {
	result := string(Encoder(common.RespValue{
		Type:   enums.BulkStringRespType,
		IsNull: true,
	}))
	if result != "$-1\r\n" {
		t.Fatalf("expected $-1 got %q", result)
	}
}

// Empty string is NOT null
func TestEncodeBulkStringEmpty(t *testing.T) {
	result := string(Encoder(common.RespValue{
		Type: enums.BulkStringRespType,
		Str:  "",
	}))
	if result != "$0\r\n\r\n" {
		t.Fatalf("expected $0 got %q", result)
	}
}
