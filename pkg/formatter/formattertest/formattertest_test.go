package formattertest

import (
	"errors"
	"testing"

	"github.com/motchang/marid/pkg/formatter"
)

func TestMockFormatterDefaults(t *testing.T) {
	m := &MockFormatter{}

	if got := m.Name(); got != "mock" {
		t.Fatalf("expected default name %q, got %q", "mock", got)
	}

	if got := m.MediaType(); got != "text/plain" {
		t.Fatalf("expected default media type %q, got %q", "text/plain", got)
	}
}

func TestMockFormatterConfiguredValues(t *testing.T) {
	m := &MockFormatter{NameValue: "custom", MediaTypeValue: "text/custom"}

	if got := m.Name(); got != "custom" {
		t.Fatalf("expected configured name %q, got %q", "custom", got)
	}

	if got := m.MediaType(); got != "text/custom" {
		t.Fatalf("expected configured media type %q, got %q", "text/custom", got)
	}
}

func TestMockFormatterRenderWithoutFunc(t *testing.T) {
	m := &MockFormatter{}

	data := SampleRenderData()
	_, err := m.Render(data)
	if err == nil {
		t.Fatal("expected error when RenderFunc is not provided")
	}

	if len(m.RenderCalls) != 1 {
		t.Fatalf("expected RenderCalls to record 1 call, got %d", len(m.RenderCalls))
	}
}

func TestMockFormatterRenderDelegates(t *testing.T) {
	renderErr := errors.New("render failure")
	m := &MockFormatter{
		RenderFunc: func(data formatter.RenderData) (string, error) {
			return "", renderErr
		},
	}

	first := SampleRenderData()
	if _, err := m.Render(first); !errors.Is(err, renderErr) {
		t.Fatalf("expected delegated error, got %v", err)
	}

	second := formatter.RenderData{}
	m.RenderFunc = func(data formatter.RenderData) (string, error) {
		return "rendered", nil
	}
	out, err := m.Render(second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "rendered" {
		t.Fatalf("expected %q, got %q", "rendered", out)
	}

	if len(m.RenderCalls) != 2 {
		t.Fatalf("expected RenderCalls to record 2 calls, got %d", len(m.RenderCalls))
	}
}
