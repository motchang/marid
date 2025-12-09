package mermaid

import (
	"strings"
	"testing"

	"github.com/motchang/marid/pkg/formatter"
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

	data := formatter.RenderData{
		Tables: []formatter.Table{
			{
				Name:       "teams",
				PrimaryKey: []string{"id"},
				Columns: []formatter.Column{
					{Name: "id", DataType: "int"},
					{Name: "name", DataType: "text"},
				},
			},
			{
				Name:       "users",
				PrimaryKey: []string{"id"},
				Columns: []formatter.Column{
					{Name: "id", DataType: "int"},
					{Name: "email", DataType: "varchar", IsUnique: true},
					{Name: "team_id", DataType: "int"},
				},
				ForeignKeys: []formatter.ForeignKey{
					{
						ColumnName:       "team_id",
						ReferencedTable:  "teams",
						ReferencedColumn: "id",
						RelationName:     "belongs_to",
					},
				},
			},
		},
	}

	got, err := f.Render(data)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	const want = `erDiagram
    teams {
        id int PK
        name text
    }
    users {
        id int PK
        email varchar UK
        team_id int FK
    }
    teams ||--o{ users : "belongs_to"
`

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
