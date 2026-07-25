package mermaid

import (
	"fmt"
	"sort"
	"strings"

	"github.com/motchang/marid/pkg/formatter"
)

func init() {
	formatter.Register("mermaid", func() formatter.Formatter {
		return New()
	})
}

// Formatter renders ER diagrams using Mermaid syntax.
type Formatter struct{}

// New creates a new Mermaid formatter instance.
func New() Formatter {
	return Formatter{}
}

// Name returns the formatter name.
func (f Formatter) Name() string {
	return "mermaid"
}

// MediaType returns the formatter output media type.
func (f Formatter) MediaType() string {
	return "text/plain"
}

// Render builds a Mermaid ER diagram from the provided render data.
func (f Formatter) Render(data formatter.RenderData) (string, error) {
	if len(data.Tables) == 0 {
		return "", fmt.Errorf("no tables found in schema")
	}

	var builder strings.Builder

	builder.WriteString("erDiagram\n")
	writeTables(&builder, data.Tables)
	writeRelationships(&builder, buildRelationships(data.Tables))

	return builder.String(), nil
}

func writeTables(builder *strings.Builder, tables []formatter.Table) {
	for _, table := range tables {
		_, _ = fmt.Fprintf(builder, "    %s {\n", table.Name)

		for _, column := range table.Columns {
			builder.WriteString(columnAttrLine(table, column) + "\n")
		}

		builder.WriteString("    }\n")
	}
}

func columnAttrLine(table formatter.Table, column formatter.Column) string {
	attrLine := fmt.Sprintf("        %s %s", column.Name, column.DataType)

	if keyConstraints := columnKeyConstraints(table, column); len(keyConstraints) > 0 {
		attrLine += " " + strings.Join(keyConstraints, ", ")
	}

	if column.Comment != "" {
		attrLine += fmt.Sprintf(" \"%s\"", column.Comment)
	}

	return attrLine
}

func columnKeyConstraints(table formatter.Table, column formatter.Column) []string {
	keyConstraints := []string{}

	isPrimary := contains(table.PrimaryKey, column.Name)
	if isPrimary {
		keyConstraints = append(keyConstraints, "PK")
	}

	for _, fk := range table.ForeignKeys {
		if fk.ColumnName == column.Name {
			keyConstraints = append(keyConstraints, "FK")
			break
		}
	}

	if column.IsUnique && !isPrimary {
		keyConstraints = append(keyConstraints, "UK")
	}

	return keyConstraints
}

// relationship is a foreign-key edge rendered as a Mermaid ER relationship line.
type relationship struct {
	SourceTable      string
	TargetTable      string
	RelationName     string
	CrossingDistance int
}

// buildRelationships collects every foreign key across tables as a
// relationship, sorted by crossing distance so closely related tables are
// rendered near each other.
func buildRelationships(tables []formatter.Table) []relationship {
	tablePositions := make(map[string]int)
	for i, table := range tables {
		tablePositions[table.Name] = i
	}

	var relationships []relationship
	for _, table := range tables {
		targetPos, targetExists := tablePositions[table.Name]

		for _, fk := range table.ForeignKeys {
			relationships = append(relationships, relationship{
				SourceTable:      fk.ReferencedTable,
				TargetTable:      table.Name,
				RelationName:     fk.RelationName,
				CrossingDistance: crossingDistance(tablePositions, fk.ReferencedTable, targetPos, targetExists),
			})
		}
	}

	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].CrossingDistance < relationships[j].CrossingDistance
	})

	return relationships
}

func crossingDistance(tablePositions map[string]int, sourceTable string, targetPos int, targetExists bool) int {
	sourcePos, sourceExists := tablePositions[sourceTable]
	if !sourceExists || !targetExists {
		return 0
	}
	return abs(sourcePos - targetPos)
}

func writeRelationships(builder *strings.Builder, relationships []relationship) {
	for _, rel := range relationships {
		_, _ = fmt.Fprintf(builder, "    %s ||--o{ %s : \"%s\"\n",
			rel.SourceTable,
			rel.TargetTable,
			rel.RelationName)
	}
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
