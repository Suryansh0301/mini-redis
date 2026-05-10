package datastore

import (
	"testing"

	"github.com/suryansh0301/mini-redis/internal/core/commands"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func makeCommand(name string, args ...string) commands.Command {
	return commands.Command{Name: name, Args: args}
}

// ── Unknown Command ───────────────────────────────────────────────

func TestExecuteUnknownCommand(t *testing.T) {
	exec := NewExecutor()
	resp := exec.Execute(makeCommand("INVALID"))
	if resp.Type != enums.ErrorRespType {
		t.Fatalf("expected error type got %+v", resp)
	}
	expected := "ERR unknown command 'INVALID'"
	if resp.Str != expected {
		t.Fatalf("expected %q got %q", expected, resp.Str)
	}
}

// ── PING ──────────────────────────────────────────────────────────

func TestExecutePing(t *testing.T) {
	exec := NewExecutor()
	resp := exec.Execute(makeCommand("PING"))
	if resp.Type != enums.SimpleStringRespType || resp.Str != "PONG" {
		t.Fatalf("expected PONG got %+v", resp)
	}
}

// ── SET / GET ─────────────────────────────────────────────────────

func TestExecuteSetGet(t *testing.T) {
	exec := NewExecutor()

	// SET
	setResp := exec.Execute(makeCommand("SET", "foo", "bar"))
	if setResp.Type != enums.SimpleStringRespType || setResp.Str != "OK" {
		t.Fatalf("expected OK got %+v", setResp)
	}

	// GET existing
	getResp := exec.Execute(makeCommand("GET", "foo"))
	if getResp.Type != enums.BulkStringRespType || getResp.Str != "bar" {
		t.Fatalf("expected bar got %+v", getResp)
	}

	// GET missing
	getMissing := exec.Execute(makeCommand("GET", "missing"))
	if !getMissing.IsNull {
		t.Fatalf("expected null bulk string got %+v", getMissing)
	}
}

// ── INCR ──────────────────────────────────────────────────────────

func TestExecuteIncr(t *testing.T) {
	exec := NewExecutor()

	// INCR missing key — should start from 0
	resp := exec.Execute(makeCommand("INCR", "counter"))
	if resp.Type != enums.IntRespType || resp.Int != 1 {
		t.Fatalf("expected 1 got %+v", resp)
	}

	// INCR again
	resp = exec.Execute(makeCommand("INCR", "counter"))
	if resp.Int != 2 {
		t.Fatalf("expected 2 got %+v", resp)
	}

	// INCR non integer
	exec.Execute(makeCommand("SET", "foo", "bar"))
	resp = exec.Execute(makeCommand("INCR", "foo"))
	if resp.Type != enums.ErrorRespType {
		t.Fatalf("expected error got %+v", resp)
	}
}

// ── DEL ───────────────────────────────────────────────────────────

func TestExecuteDel(t *testing.T) {
	exec := NewExecutor()

	exec.Execute(makeCommand("SET", "foo", "bar"))

	// DEL existing
	resp := exec.Execute(makeCommand("DEL", "foo"))
	if resp.Type != enums.IntRespType || resp.Int != 1 {
		t.Fatalf("expected 1 got %+v", resp)
	}

	// Verify deleted
	getResp := exec.Execute(makeCommand("GET", "foo"))
	if !getResp.IsNull {
		t.Fatal("key should be deleted")
	}

	// DEL missing
	resp = exec.Execute(makeCommand("DEL", "foo"))
	if resp.Int != 0 {
		t.Fatalf("expected 0 got %+v", resp)
	}
}

// ── State isolation ───────────────────────────────────────────────

func TestExecutorStateIsolation(t *testing.T) {
	// Two executors should have independent datastores
	exec1 := NewExecutor()
	exec2 := NewExecutor()

	exec1.Execute(makeCommand("SET", "foo", "bar"))

	resp := exec2.Execute(makeCommand("GET", "foo"))
	if !resp.IsNull {
		t.Fatal("exec2 should not see exec1 data")
	}
}
