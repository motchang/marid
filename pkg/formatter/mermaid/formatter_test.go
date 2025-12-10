package mermaid

import (
	"strings"
	"testing"

	"github.com/motchang/marid/pkg/formatter"
	"github.com/motchang/marid/pkg/formatter/formattertest"
)

func TestFormatterMetadata(t *testing.T) {
	f := New()

	if got, want := f.Name(), "mermaid"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}

	if got, want := f.MediaType(), "text/plain"; got != want {
		t.Fatalf("MediaType() = %q, want %q", got, want)
	}
}

func TestRenderNoTables(t *testing.T) {
	f := New()

	if _, err := f.Render(formatter.RenderData{}); err == nil {
		t.Fatalf("Render should fail when no tables are provided")
	}
}

func TestRenderGeneratesMermaidDiagram(t *testing.T) {
	f := New()

	got, err := f.Render(formattertest.SampleRenderData())
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := formattertest.SampleMermaidOutput()

	if got != want {
		t.Fatalf("Render() output mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRenderOrdersRelationshipsByCrossingDistance(t *testing.T) {
	f := New()

	data := formatter.RenderData{
		Tables: []formatter.Table{
			{ // index 0
				Name:       "users",
				PrimaryKey: []string{"id"},
				Columns:    []formatter.Column{{Name: "id", DataType: "int"}},
			},
			{ // index 1
				Name:       "posts",
				PrimaryKey: []string{"id"},
				Columns:    []formatter.Column{{Name: "id", DataType: "int"}},
			},
			{ // index 2
				Name:       "comments",
				PrimaryKey: []string{"id"},
				Columns: []formatter.Column{
					{Name: "id", DataType: "int"},
					{Name: "user_id", DataType: "int"},
					{Name: "post_id", DataType: "int"},
				},
				ForeignKeys: []formatter.ForeignKey{
					{
						ColumnName:       "user_id",
						ReferencedTable:  "users",
						ReferencedColumn: "id",
						RelationName:     "comment_author",
					},
					{
						ColumnName:       "post_id",
						ReferencedTable:  "posts",
						ReferencedColumn: "id",
						RelationName:     "comment_post",
					},
				},
			},
		},
	}

	got, err := f.Render(data)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	firstRel := "    posts ||--o{ comments : \"comment_post\"\n"
	secondRel := "    users ||--o{ comments : \"comment_author\"\n"

	idx1 := strings.Index(got, firstRel)
	idx2 := strings.Index(got, secondRel)

	if idx1 == -1 || idx2 == -1 {
		t.Fatalf("expected both relationships in output\n%s", got)
	}

	if idx1 > idx2 {
		t.Fatalf("relationships not ordered by crossing distance\noutput: %s", got)
	}
}

func TestRenderIncludesColumnComments(t *testing.T) {
	f := New()

	data := formatter.RenderData{
		Tables: []formatter.Table{
			{
				Name:       "notes",
				PrimaryKey: []string{"id"},
				Columns: []formatter.Column{
					{Name: "id", DataType: "int"},
					{Name: "content", DataType: "text", Comment: "freeform notes"},
				},
			},
		},
	}

	got, err := f.Render(data)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if !strings.Contains(got, "        content text \"freeform notes\"\n") {
		t.Fatalf("expected column comments to be rendered, got:\n%s", got)
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		out  int
	}{
		{name: "negative", in: -3, out: 3},
		{name: "positive", in: 5, out: 5},
		{name: "zero", in: 0, out: 0},
	}

	for _, tt := range tests {
		if got := abs(tt.in); got != tt.out {
			t.Fatalf("abs(%s) = %d, want %d", tt.name, got, tt.out)
		}
	}
}
