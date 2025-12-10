package formatter_test

import (
	"testing"

	"github.com/motchang/marid/pkg/formatter"
	"github.com/motchang/marid/pkg/formatter/formattertest"
	"github.com/motchang/marid/pkg/formatter/mermaid"
)

func TestFormatterContract(t *testing.T) {
	t.Parallel()

	testData := formattertest.SampleRenderData()

	tests := []struct {
		name            string
		formatter       formatter.Formatter
		wantName        string
		wantMediaType   string
		wantRenderMatch string
	}{
		{
			name:            "mermaid implements contract",
			formatter:       mermaid.New(),
			wantName:        "mermaid",
			wantMediaType:   "text/plain",
			wantRenderMatch: formattertest.SampleMermaidOutput(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.formatter.Name(); got != tt.wantName {
				t.Fatalf("Name() = %q, want %q", got, tt.wantName)
			}

			if got := tt.formatter.MediaType(); got != tt.wantMediaType {
				t.Fatalf("MediaType() = %q, want %q", got, tt.wantMediaType)
			}

			got, err := tt.formatter.Render(testData)
			if err != nil {
				t.Fatalf("Render returned error: %v", err)
			}

			if got != tt.wantRenderMatch {
				t.Fatalf("Render() output mismatch\n--- want ---\n%s\n--- got ---\n%s", tt.wantRenderMatch, got)
			}
		})
	}
}

func TestFormatterContractRejectsEmptyTables(t *testing.T) {
	t.Parallel()

	f := mermaid.New()
	if _, err := f.Render(formatter.RenderData{}); err == nil {
		t.Fatalf("Render should fail when no tables are provided")
	}
}
