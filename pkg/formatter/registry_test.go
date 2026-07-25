package formatter_test

import (
	"strings"
	"testing"

	"github.com/motchang/marid/pkg/formatter"
	"github.com/motchang/marid/pkg/formatter/formattertest"
	_ "github.com/motchang/marid/pkg/formatter/mermaid"
)

func TestGetReturnsDefaultWhenEmpty(t *testing.T) {
	fmttr, err := formatter.Get("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fmttr.Name() != formatter.DefaultFormat {
		t.Fatalf("expected default formatter %q, got %q", formatter.DefaultFormat, fmttr.Name())
	}
}

func TestGetUnknownFormat(t *testing.T) {
	_, err := formatter.Get("unknown")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}

	if !strings.Contains(err.Error(), "unknown format \"unknown\". Available formats: ") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestAvailableIsSorted(t *testing.T) {
	names := formatter.Available()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Fatalf("expected names to be sorted, got %v", names)
		}
	}
}

func newMockFactory() formatter.Factory {
	return func() formatter.Formatter { return &formattertest.MockFormatter{} }
}

func expectPanic(t *testing.T, wantSubstring string, fn func()) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, got none", wantSubstring)
		}

		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, wantSubstring) {
			t.Fatalf("expected panic containing %q, got %v", wantSubstring, r)
		}
	}()

	fn()
}

func TestRegisterPanicsOnEmptyName(t *testing.T) {
	expectPanic(t, "formatter name cannot be empty", func() {
		formatter.Register("", newMockFactory())
	})
}

func TestRegisterPanicsOnNilFactory(t *testing.T) {
	expectPanic(t, "formatter factory cannot be nil", func() {
		formatter.Register("test-register-nil-factory", nil)
	})
}

func TestRegisterPanicsOnDuplicateName(t *testing.T) {
	const name = "test-register-duplicate"

	formatter.Register(name, newMockFactory())

	expectPanic(t, "already registered", func() {
		formatter.Register(name, newMockFactory())
	})
}
