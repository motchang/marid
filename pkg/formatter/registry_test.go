package formatter_test

import (
	"strings"
	"testing"

	"github.com/motchang/marid/pkg/formatter"
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
