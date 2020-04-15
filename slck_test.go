package slck

import (
	"fmt"
	"testing"
)

func TestRawCommand_Scan(t *testing.T) {
	kind, args := commandJoin, []byte("alice")
	src := fmt.Sprintf("%s %s", kind, args)

	var cmd rawCommand
	if _, err := fmt.Sscan(src, &cmd); err != nil {
		t.Errorf("should have scanned: %s", err)
		return
	}

	if cmd.kind != kind {
		t.Errorf("should have scanned the expected kind: %s", reportUnexpected("kind", cmd.kind, kind))
		return
	}
	if string(cmd.args) != string(args) {
		t.Errorf("should have scanned the expected args: %s", reportUnexpected("kind", string(cmd.args), string(args)))
		return
	}
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
