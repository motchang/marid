package diagram

import (
	"testing"

	"github.com/motchang/marid/internal/schema"
)

func TestGenerate(t *testing.T) {
	dbSchema := &schema.DatabaseSchema{
		Tables: []schema.Table{
			{
				Name:    "users",
				Comment: "User master",
				Columns: []schema.Column{
					{Name: "id", DataType: "int", IsPrimary: true},
					{Name: "name", DataType: "varchar", Comment: "User name"},
				},
				PrimaryKey: []string{"id"},
			},
			{
				Name: "orders",
				Columns: []schema.Column{
					{Name: "id", DataType: "int", IsPrimary: true},
					{Name: "user_id", DataType: "int"},
					{Name: "total", DataType: "decimal"},
				},
				PrimaryKey: []string{"id"},
				ForeignKeys: []schema.ForeignKey{
					{ColumnName: "user_id", ReferencedTable: "users", ReferencedColumn: "id", RelationName: "orders_users_fk"},
				},
			},
		},
	}

	expected := "erDiagram\n" +
		"    users {\n" +
		"        id int PK\n" +
		"        name varchar \"User name\"\n" +
		"    }\n" +
		"    orders {\n" +
		"        id int PK\n" +
		"        user_id int FK\n" +
		"        total decimal\n" +
		"    }\n" +
		"    users ||--o{ orders : \"orders_users_fk\"\n"

	got, err := Generate(dbSchema, "")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if got != expected {
		t.Fatalf("unexpected diagram output:\nexpected:\n%s\ngot:\n%s", expected, got)
	}
}

func TestGenerateReturnsErrorWhenNoTables(t *testing.T) {
	_, err := Generate(&schema.DatabaseSchema{}, "")
	if err == nil {
		t.Fatal("expected error when schema has no tables")
	}
}

func TestGenerateReturnsErrorForUnknownFormat(t *testing.T) {
	_, err := Generate(&schema.DatabaseSchema{Tables: []schema.Table{{Name: "users"}}}, "unknown")
	if err == nil {
		t.Fatal("expected error when format is unknown")
	}

	const want = "unknown format \"unknown\". Available formats: mermaid"
	if err.Error() != want {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}
