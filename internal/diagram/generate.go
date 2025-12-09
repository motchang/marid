package diagram

import (
	"fmt"

	"github.com/motchang/marid/internal/schema"
	"github.com/motchang/marid/pkg/formatter"
	_ "github.com/motchang/marid/pkg/formatter/mermaid"
)

// Generator coordinates rendering using a formatter.
type Generator struct {
	formatter formatter.Formatter
}

// New creates a generator that uses the provided formatter.
//
// In Go, constructor-style functions typically use the package name for
// disambiguation (e.g. diagram.New), so the bare New name is idiomatic.
func New(f formatter.Formatter) *Generator {
	return &Generator{formatter: f}
}

// Generate renders a diagram using the formatter associated with the provided format name.
//
// If format is empty, formatter.DefaultFormat is used.
func Generate(dbSchema *schema.DatabaseSchema, format string) (string, error) {
	fmttr, err := formatter.Get(format)
	if err != nil {
		return "", err
	}

	generator := New(fmttr)
	return generator.Generate(dbSchema)
}

// Generate renders a diagram using the configured formatter.
func (g *Generator) Generate(dbSchema *schema.DatabaseSchema) (string, error) {
	if dbSchema == nil || len(dbSchema.Tables) == 0 {
		return "", fmt.Errorf("no tables found in schema")
	}

	renderData := toRenderData(dbSchema)
	return g.formatter.Render(renderData)
}

func toRenderData(dbSchema *schema.DatabaseSchema) formatter.RenderData {
	tables := make([]formatter.Table, len(dbSchema.Tables))

	for i, tbl := range dbSchema.Tables {
		columns := make([]formatter.Column, len(tbl.Columns))
		for ci, col := range tbl.Columns {
			columns[ci] = formatter.Column{
				Name:       col.Name,
				DataType:   col.DataType,
				IsNullable: col.IsNullable,
				IsPrimary:  col.IsPrimary,
				IsUnique:   col.IsUnique,
				Comment:    col.Comment,
			}
		}

		foreignKeys := make([]formatter.ForeignKey, len(tbl.ForeignKeys))
		for fi, fk := range tbl.ForeignKeys {
			foreignKeys[fi] = formatter.ForeignKey{
				ColumnName:       fk.ColumnName,
				ReferencedTable:  fk.ReferencedTable,
				ReferencedColumn: fk.ReferencedColumn,
				RelationName:     fk.RelationName,
			}
		}

		tables[i] = formatter.Table{
			Name:        tbl.Name,
			Comment:     tbl.Comment,
			Columns:     columns,
			PrimaryKey:  append([]string(nil), tbl.PrimaryKey...),
			ForeignKeys: foreignKeys,
		}
	}

	return formatter.RenderData{Tables: tables}
}
