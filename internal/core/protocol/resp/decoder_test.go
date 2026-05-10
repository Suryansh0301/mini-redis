package resp

import (
	"testing"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

// Helper to build a ParseResp from a RespValue directly
func makeParseResp(v *common.RespValue) ParseResp {
	return ParseResp{
		Resp:          v,
		bytesConsumed: 1,
	}
}

func TestDecoderPing(t *testing.T) {
	input := makeParseResp(&common.RespValue{
		Type: enums.ArrayRespType,
		Array: []*common.RespValue{
			{Type: enums.BulkStringRespType, Str: "ping"},
		},
	})

	cmd, err := Decoder(input)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Name != "PING" {
		t.Fatalf("expected PING got %q", cmd.Name)
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("expected no args got %v", cmd.Args)
	}
}

func TestDecoderSet(t *testing.T) {
	input := makeParseResp(&common.RespValue{
		Type: enums.ArrayRespType,
		Array: []*common.RespValue{
			{Type: enums.BulkStringRespType, Str: "SET"},
			{Type: enums.BulkStringRespType, Str: "foo"},
			{Type: enums.BulkStringRespType, Str: "bar"},
		},
	})

	cmd, err := Decoder(input)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Name != "SET" {
		t.Fatalf("expected SET got %q", cmd.Name)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "foo" || cmd.Args[1] != "bar" {
		t.Fatalf("unexpected args %v", cmd.Args)
	}
}

func TestDecoderCaseInsensitive(t *testing.T) {
	// Redis commands are case insensitive
	tests := []string{"ping", "PING", "Ping", "pInG"}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			input := makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{Type: enums.BulkStringRespType, Str: name},
				},
			})
			cmd, err := Decoder(input)
			if err != nil {
				t.Fatal(err)
			}
			if cmd.Name != "PING" {
				t.Fatalf("expected PING got %q", cmd.Name)
			}
		})
	}
}

func TestDecoderErrors(t *testing.T) {
	tests := []struct {
		name  string
		input ParseResp
	}{
		{
			name:  "nil resp",
			input: ParseResp{Resp: nil},
		},
		{
			name: "not an array",
			input: makeParseResp(&common.RespValue{
				Type: enums.SimpleStringRespType,
				Str:  "PING",
			}),
		},
		{
			name: "null array",
			input: makeParseResp(&common.RespValue{
				Type:   enums.ArrayRespType,
				IsNull: true,
			}),
		},
		{
			name: "empty array",
			input: makeParseResp(&common.RespValue{
				Type:  enums.ArrayRespType,
				Array: []*common.RespValue{},
			}),
		},
		{
			name: "non bulk string command name",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{Type: enums.SimpleStringRespType, Str: "PING"},
				},
			}),
		},
		{
			name: "null command name",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{Type: enums.BulkStringRespType, IsNull: true},
				},
			}),
		},
		{
			name: "null argument",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{Type: enums.BulkStringRespType, Str: "GET"},
					{Type: enums.BulkStringRespType, IsNull: true},
				},
			}),
		},
		{
			name: "non bulk string argument",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{Type: enums.BulkStringRespType, Str: "GET"},
					{Type: enums.SimpleStringRespType, Str: "foo"},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decoder(tt.input)
			if err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}
