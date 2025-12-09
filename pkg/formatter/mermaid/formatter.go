package mermaid

import (
	"fmt"
	"sort"
	"strings"

	"github.com/motchang/marid/pkg/formatter"
)

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

	for _, table := range data.Tables {

		builder.WriteString(fmt.Sprintf("    %s {\n", table.Name))

		for _, column := range table.Columns {
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

			attrLine := fmt.Sprintf("        %s %s", column.Name, column.DataType)

			if len(keyConstraints) > 0 {
				attrLine += " " + strings.Join(keyConstraints, ", ")
			}

			if column.Comment != "" {
				attrLine += fmt.Sprintf(" \"%s\"", column.Comment)
			}

			builder.WriteString(attrLine + "\n")
		}
		builder.WriteString("    }\n")
	}

	type relationship struct {
		SourceTable      string
		TargetTable      string
		RelationName     string
		CrossingDistance int
	}

	var relationships []relationship

	tablePositions := make(map[string]int)
	for i, table := range data.Tables {
		tablePositions[table.Name] = i
	}

	for _, table := range data.Tables {
		for _, fk := range table.ForeignKeys {
			sourcePos, sourceExists := tablePositions[fk.ReferencedTable]
			targetPos, targetExists := tablePositions[table.Name]

			crossingDistance := 0
			if sourceExists && targetExists {
				crossingDistance = abs(sourcePos - targetPos)
			}

			relationships = append(relationships, relationship{
				SourceTable:      fk.ReferencedTable,
				TargetTable:      table.Name,
				RelationName:     fk.RelationName,
				CrossingDistance: crossingDistance,
			})
		}
	}

	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].CrossingDistance < relationships[j].CrossingDistance
	})

	for _, rel := range relationships {
		builder.WriteString(fmt.Sprintf("    %s ||--o{ %s : \"%s\"\n",
			rel.SourceTable,
			rel.TargetTable,
			rel.RelationName))
	}

	return builder.String(), nil
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
