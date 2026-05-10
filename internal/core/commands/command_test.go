package commands

import (
	"testing"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

// ── Helpers ──────────────────────────────────────────────────────

func makeStore(pairs ...string) map[string]string {
	store := make(map[string]string)
	for i := 0; i+1 < len(pairs); i += 2 {
		store[pairs[i]] = pairs[i+1]
	}
	return store
}

func assertOK(t *testing.T, resp common.RespValue) {
	t.Helper()
	if resp.Type != enums.SimpleStringRespType || resp.Str != "OK" {
		t.Fatalf("expected +OK got %+v", resp)
	}
}

func assertError(t *testing.T, resp common.RespValue, expectedMsg string) {
	t.Helper()
	if resp.Type != enums.ErrorRespType {
		t.Fatalf("expected error type got %+v", resp)
	}
	if resp.Str != expectedMsg {
		t.Fatalf("expected error %q got %q", expectedMsg, resp.Str)
	}
}

func assertBulkString(t *testing.T, resp common.RespValue, expected string) {
	t.Helper()
	if resp.Type != enums.BulkStringRespType {
		t.Fatalf("expected bulk string got %+v", resp)
	}
	if resp.Str != expected {
		t.Fatalf("expected %q got %q", expected, resp.Str)
	}
}

func assertNullBulk(t *testing.T, resp common.RespValue) {
	t.Helper()
	if resp.Type != enums.BulkStringRespType || !resp.IsNull {
		t.Fatalf("expected null bulk string got %+v", resp)
	}
}

func assertInteger(t *testing.T, resp common.RespValue, expected int64) {
	t.Helper()
	if resp.Type != enums.IntRespType {
		t.Fatalf("expected integer got %+v", resp)
	}
	if resp.Int != expected {
		t.Fatalf("expected %d got %d", expected, resp.Int)
	}
}

// ── PING ─────────────────────────────────────────────────────────

func TestPing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no args",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "with args",
			args:        []string{"hello"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "PING", Args: tt.args}
			resp := HandlerPing(cmd, nil)
			if tt.expectError {
				assertError(t, resp, common.WrongNumberOfArgumentsError("PING"))
			} else {
				if resp.Type != enums.SimpleStringRespType || resp.Str != "PONG" {
					t.Fatalf("expected PONG got %+v", resp)
				}
			}
		})
	}
}

// ── ECHO ─────────────────────────────────────────────────────────

func TestEcho(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expected    string
	}{
		{
			name:     "valid",
			args:     []string{"hello"},
			expected: "hello",
		},
		{
			name:        "no args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"hello", "world"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "ECHO", Args: tt.args}
			resp := HandlerEcho(cmd, nil)
			if tt.expectError {
				assertError(t, resp, common.WrongNumberOfArgumentsError("ECHO"))
			} else {
				assertBulkString(t, resp, tt.expected)
			}
		})
	}
}

// ── SET ──────────────────────────────────────────────────────────

func TestSet(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name: "valid",
			args: []string{"foo", "bar"},
		},
		{
			name:        "no args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "only key",
			args:        []string{"foo"},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"foo", "bar", "baz"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := makeStore()
			cmd := Command{Name: "SET", Args: tt.args}
			resp := HandlerSet(cmd, store)
			if tt.expectError {
				assertError(t, resp, common.WrongNumberOfArgumentsError("SET"))
			} else {
				assertOK(t, resp)
				if store[tt.args[0]] != tt.args[1] {
					t.Fatalf("store not updated, expected %q got %q", tt.args[1], store[tt.args[0]])
				}
			}
		})
	}
}

// ── GET ──────────────────────────────────────────────────────────

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]string
		args        []string
		expectNull  bool
		expectError bool
		expected    string
	}{
		{
			name:     "existing key",
			store:    makeStore("foo", "bar"),
			args:     []string{"foo"},
			expected: "bar",
		},
		{
			name:       "missing key",
			store:      makeStore(),
			args:       []string{"foo"},
			expectNull: true,
		},
		{
			name:        "no args",
			store:       makeStore(),
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			store:       makeStore(),
			args:        []string{"foo", "bar"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "GET", Args: tt.args}
			resp := HandlerGet(cmd, tt.store)
			if tt.expectError {
				assertError(t, resp, common.WrongNumberOfArgumentsError("GET"))
			} else if tt.expectNull {
				assertNullBulk(t, resp)
			} else {
				assertBulkString(t, resp, tt.expected)
			}
		})
	}
}

// ── INCR ─────────────────────────────────────────────────────────

func TestIncr(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]string
		args        []string
		expectError bool
		errorMsg    string
		expected    int64
	}{
		{
			name:     "existing integer key",
			store:    makeStore("counter", "5"),
			args:     []string{"counter"},
			expected: 6,
		},
		{
			name:     "missing key starts at 0",
			store:    makeStore(),
			args:     []string{"counter"},
			expected: 1,
		},
		{
			name:        "non integer value",
			store:       makeStore("foo", "bar"),
			args:        []string{"foo"},
			expectError: true,
			errorMsg:    "ERR value is not an integer or out of range",
		},
		{
			name:        "no args",
			store:       makeStore(),
			args:        []string{},
			expectError: true,
			errorMsg:    common.WrongNumberOfArgumentsError("INCR"),
		},
		{
			name:        "too many args",
			store:       makeStore(),
			args:        []string{"foo", "bar"},
			expectError: true,
			errorMsg:    common.WrongNumberOfArgumentsError("INCR"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "INCR", Args: tt.args}
			resp := HandlerIncr(cmd, tt.store)
			if tt.expectError {
				assertError(t, resp, tt.errorMsg)
			} else {
				assertInteger(t, resp, tt.expected)
			}
		})
	}
}

// ── DEL ──────────────────────────────────────────────────────────

func TestDel(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]string
		args        []string
		expectError bool
		expected    int64
	}{
		{
			name:     "existing key",
			store:    makeStore("foo", "bar"),
			args:     []string{"foo"},
			expected: 1,
		},
		{
			name:     "missing key",
			store:    makeStore(),
			args:     []string{"foo"},
			expected: 0,
		},
		{
			name:        "no args",
			store:       makeStore(),
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			store:       makeStore(),
			args:        []string{"foo", "bar"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "DEL", Args: tt.args}
			resp := HandlerDel(cmd, tt.store)
			if tt.expectError {
				assertError(t, resp, common.WrongNumberOfArgumentsError("DEL"))
			} else {
				assertInteger(t, resp, tt.expected)
				if tt.expected == 1 {
					if _, exists := tt.store[tt.args[0]]; exists {
						t.Fatal("key should have been deleted from store")
					}
				}
			}
		})
	}
}
