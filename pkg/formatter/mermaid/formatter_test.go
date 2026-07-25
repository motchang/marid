package mermaid

import (
	"slices"
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

// TestInitRegistersFormatter exercises the init() self-registration that lets
// --format mermaid resolve through the registry with no CLI wiring, and checks
// the registered factory hands back a usable formatter.
func TestInitRegistersFormatter(t *testing.T) {
	got, err := formatter.Get("mermaid")
	if err != nil {
		t.Fatalf("formatter.Get(%q) returned error: %v", "mermaid", err)
	}

	if _, ok := got.(Formatter); !ok {
		t.Fatalf("formatter.Get(%q) returned %T, want mermaid.Formatter", "mermaid", got)
	}

	if name := got.Name(); name != "mermaid" {
		t.Errorf("registered formatter Name() = %q, want %q", name, "mermaid")
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

// TestRenderForeignKeyToTableOutsideDiagram covers what marid produces when
// --tables filters the schema: a foreign key can point at a table that is not
// part of the rendered set, so its crossing distance cannot be computed.
func TestRenderForeignKeyToTableOutsideDiagram(t *testing.T) {
	f := New()

	data := formatter.RenderData{
		Tables: []formatter.Table{
			{
				Name:       "orders",
				PrimaryKey: []string{"id"},
				Columns: []formatter.Column{
					{Name: "id", DataType: "int"},
					{Name: "customer_id", DataType: "int"},
				},
				ForeignKeys: []formatter.ForeignKey{
					{
						ColumnName:       "customer_id",
						ReferencedTable:  "customers", // filtered out of the diagram
						ReferencedColumn: "id",
						RelationName:     "placed_by",
					},
				},
			},
		},
	}

	got, err := f.Render(data)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	// The relationship is still emitted, naming the table that is absent.
	if want := "    customers ||--o{ orders : \"placed_by\"\n"; !strings.Contains(got, want) {
		t.Errorf("expected relationship to a table outside the diagram\nwant substring: %q\ngot:\n%s", want, got)
	}

	// The column keeps its FK marker even though the target is not rendered.
	if want := "        customer_id int FK\n"; !strings.Contains(got, want) {
		t.Errorf("expected customer_id to be marked FK\nwant substring: %q\ngot:\n%s", want, got)
	}
}

func TestRenderTableWithoutColumns(t *testing.T) {
	f := New()

	got, err := f.Render(formatter.RenderData{Tables: []formatter.Table{{Name: "empty"}}})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if want := "erDiagram\n    empty {\n    }\n"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestColumnKeyConstraints(t *testing.T) {
	tests := []struct {
		name   string
		table  formatter.Table
		column formatter.Column
		want   []string
	}{
		{
			name:   "no constraints",
			table:  formatter.Table{Name: "posts"},
			column: formatter.Column{Name: "title", DataType: "varchar"},
			want:   []string{},
		},
		{
			name:   "primary key",
			table:  formatter.Table{Name: "posts", PrimaryKey: []string{"id"}},
			column: formatter.Column{Name: "id", DataType: "int"},
			want:   []string{"PK"},
		},
		{
			name:   "member of a composite primary key",
			table:  formatter.Table{Name: "post_tags", PrimaryKey: []string{"post_id", "tag_id"}},
			column: formatter.Column{Name: "tag_id", DataType: "int"},
			want:   []string{"PK"},
		},
		{
			name: "foreign key",
			table: formatter.Table{
				Name:        "posts",
				ForeignKeys: []formatter.ForeignKey{{ColumnName: "user_id", ReferencedTable: "users"}},
			},
			column: formatter.Column{Name: "user_id", DataType: "int"},
			want:   []string{"FK"},
		},
		{
			name: "foreign key on another column is ignored",
			table: formatter.Table{
				Name:        "posts",
				ForeignKeys: []formatter.ForeignKey{{ColumnName: "team_id", ReferencedTable: "teams"}},
			},
			column: formatter.Column{Name: "user_id", DataType: "int"},
			want:   []string{},
		},
		{
			name: "two foreign keys on the same column yield a single FK",
			table: formatter.Table{
				Name: "posts",
				ForeignKeys: []formatter.ForeignKey{
					{ColumnName: "user_id", ReferencedTable: "users"},
					{ColumnName: "user_id", ReferencedTable: "authors"},
				},
			},
			column: formatter.Column{Name: "user_id", DataType: "int"},
			want:   []string{"FK"},
		},
		{
			name:   "unique key",
			table:  formatter.Table{Name: "users"},
			column: formatter.Column{Name: "email", DataType: "varchar", IsUnique: true},
			want:   []string{"UK"},
		},
		{
			name:   "primary key suppresses the redundant unique marker",
			table:  formatter.Table{Name: "users", PrimaryKey: []string{"id"}},
			column: formatter.Column{Name: "id", DataType: "int", IsUnique: true},
			want:   []string{"PK"},
		},
		{
			name: "primary key that is also a foreign key",
			table: formatter.Table{
				Name:        "profiles",
				PrimaryKey:  []string{"user_id"},
				ForeignKeys: []formatter.ForeignKey{{ColumnName: "user_id", ReferencedTable: "users"}},
			},
			column: formatter.Column{Name: "user_id", DataType: "int"},
			want:   []string{"PK", "FK"},
		},
		{
			name: "unique foreign key keeps both markers in order",
			table: formatter.Table{
				Name:        "profiles",
				ForeignKeys: []formatter.ForeignKey{{ColumnName: "user_id", ReferencedTable: "users"}},
			},
			column: formatter.Column{Name: "user_id", DataType: "int", IsUnique: true},
			want:   []string{"FK", "UK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := columnKeyConstraints(tt.table, tt.column); !slices.Equal(got, tt.want) {
				t.Errorf("columnKeyConstraints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnAttrLine(t *testing.T) {
	tests := []struct {
		name   string
		table  formatter.Table
		column formatter.Column
		want   string
	}{
		{
			name:   "name and type only",
			table:  formatter.Table{Name: "posts"},
			column: formatter.Column{Name: "title", DataType: "varchar"},
			want:   "        title varchar",
		},
		{
			name:   "single constraint",
			table:  formatter.Table{Name: "posts", PrimaryKey: []string{"id"}},
			column: formatter.Column{Name: "id", DataType: "int"},
			want:   "        id int PK",
		},
		{
			name: "multiple constraints are comma separated",
			table: formatter.Table{
				Name:        "profiles",
				PrimaryKey:  []string{"user_id"},
				ForeignKeys: []formatter.ForeignKey{{ColumnName: "user_id", ReferencedTable: "users"}},
			},
			column: formatter.Column{Name: "user_id", DataType: "int"},
			want:   "        user_id int PK, FK",
		},
		{
			name:   "comment without constraints",
			table:  formatter.Table{Name: "posts"},
			column: formatter.Column{Name: "body", DataType: "text", Comment: "free text"},
			want:   "        body text \"free text\"",
		},
		{
			name:   "constraint and comment together",
			table:  formatter.Table{Name: "posts", PrimaryKey: []string{"id"}},
			column: formatter.Column{Name: "id", DataType: "int", Comment: "identifier"},
			want:   "        id int PK \"identifier\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := columnAttrLine(tt.table, tt.column); got != tt.want {
				t.Errorf("columnAttrLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildRelationshipsSortsByCrossingDistance(t *testing.T) {
	tables := []formatter.Table{
		{Name: "users"}, // position 0
		{Name: "posts"}, // position 1
		{Name: "tags"},  // position 2
		{
			Name: "comments", // position 3
			ForeignKeys: []formatter.ForeignKey{
				{ColumnName: "user_id", ReferencedTable: "users", RelationName: "far"},
				{ColumnName: "post_id", ReferencedTable: "posts", RelationName: "mid"},
				{ColumnName: "tag_id", ReferencedTable: "tags", RelationName: "near"},
			},
		},
	}

	got := buildRelationships(tables)

	want := []relationship{
		{SourceTable: "tags", TargetTable: "comments", RelationName: "near", CrossingDistance: 1},
		{SourceTable: "posts", TargetTable: "comments", RelationName: "mid", CrossingDistance: 2},
		{SourceTable: "users", TargetTable: "comments", RelationName: "far", CrossingDistance: 3},
	}

	if !slices.Equal(got, want) {
		t.Errorf("buildRelationships() = %+v, want %+v", got, want)
	}
}

// TestBuildRelationshipsTreatsUnknownReferencedTableAsAdjacent pins the
// consequence of an unresolvable crossing distance: it defaults to 0, so a
// relationship pointing outside the diagram sorts ahead of resolvable ones.
func TestBuildRelationshipsTreatsUnknownReferencedTableAsAdjacent(t *testing.T) {
	tables := []formatter.Table{
		{Name: "users"}, // position 0
		{
			Name: "orders", // position 1
			ForeignKeys: []formatter.ForeignKey{
				{ColumnName: "user_id", ReferencedTable: "users", RelationName: "inside"},
				{ColumnName: "customer_id", ReferencedTable: "customers", RelationName: "outside"},
			},
		},
	}

	got := buildRelationships(tables)

	want := []relationship{
		{SourceTable: "customers", TargetTable: "orders", RelationName: "outside", CrossingDistance: 0},
		{SourceTable: "users", TargetTable: "orders", RelationName: "inside", CrossingDistance: 1},
	}

	if !slices.Equal(got, want) {
		t.Errorf("buildRelationships() = %+v, want %+v", got, want)
	}
}

func TestBuildRelationshipsWithoutForeignKeys(t *testing.T) {
	got := buildRelationships([]formatter.Table{{Name: "users"}, {Name: "posts"}})

	if len(got) != 0 {
		t.Errorf("buildRelationships() = %+v, want no relationships", got)
	}
}

func TestCrossingDistance(t *testing.T) {
	positions := map[string]int{"users": 0, "posts": 1, "comments": 4}

	tests := []struct {
		name         string
		sourceTable  string
		targetPos    int
		targetExists bool
		want         int
	}{
		{name: "source after target", sourceTable: "comments", targetPos: 1, targetExists: true, want: 3},
		{name: "source before target", sourceTable: "users", targetPos: 4, targetExists: true, want: 4},
		{name: "same position", sourceTable: "posts", targetPos: 1, targetExists: true, want: 0},
		{name: "source not in the diagram", sourceTable: "absent", targetPos: 1, targetExists: true, want: 0},
		{name: "target not in the diagram", sourceTable: "users", targetPos: 0, targetExists: false, want: 0},
		{name: "neither side in the diagram", sourceTable: "absent", targetPos: 0, targetExists: false, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := crossingDistance(positions, tt.sourceTable, tt.targetPos, tt.targetExists)
			if got != tt.want {
				t.Errorf("crossingDistance(%q, %d, %t) = %d, want %d",
					tt.sourceTable, tt.targetPos, tt.targetExists, got, tt.want)
			}
		})
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
