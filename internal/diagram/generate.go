package diagram

import (
	"fmt"
	"sort"
	"strings"

	"github.com/motchang/marid/internal/schema"
)

// Generate creates a Mermaid ER diagram from the database schema
func Generate(dbSchema *schema.DatabaseSchema) (string, error) {
	if len(dbSchema.Tables) == 0 {
		return "", fmt.Errorf("no tables found in schema")
	}

	var builder strings.Builder

	// Start Mermaid ER diagram
	builder.WriteString("erDiagram\n")

	// Process each table
	for _, table := range dbSchema.Tables {
		// Add table entity
		builder.WriteString(fmt.Sprintf("    %s {\n", table.Name))

		// Add columns
		for _, column := range table.Columns {
			// Determine key constraints
			keyConstraints := []string{}

			// Check for primary key
			isPrimary := false
			for _, pk := range table.PrimaryKey {
				if pk == column.Name {
					isPrimary = true
					keyConstraints = append(keyConstraints, "PK")
					break
				}
			}

			// Check for foreign key
			for _, fk := range table.ForeignKeys {
				if fk.ColumnName == column.Name {
					keyConstraints = append(keyConstraints, "FK")
					break
				}
			}

			// Check for unique key (if column has unique constraint)
			if column.IsUnique && !isPrimary { // Primary keys are implicitly unique
				keyConstraints = append(keyConstraints, "UK")
			}

			// Format data type
			dataType := column.DataType

			// Build the attribute line
			attrLine := fmt.Sprintf("        %s %s", column.Name, dataType)

			// Add key constraints if any
			if len(keyConstraints) > 0 {
				attrLine += " " + strings.Join(keyConstraints, ", ")
			}

			// Add column comment if available
			if column.Comment != "" {
				attrLine += fmt.Sprintf(" \"%s\"", column.Comment)
			}

			builder.WriteString(attrLine + "\n")
		}
		builder.WriteString("    }\n")
	}

	// Collect all relationships
	type Relationship struct {
		SourceTable      string
		TargetTable      string
		RelationName     string
		CrossingDistance int // Measure of how far apart the tables are
	}

	var relationships []Relationship

	// Build table position map (for calculating crossing distance)
	tablePositions := make(map[string]int)
	for i, table := range dbSchema.Tables {
		tablePositions[table.Name] = i
	}

	// Collect all relationships and calculate crossing distances
	for _, table := range dbSchema.Tables {
		for _, fk := range table.ForeignKeys {
			sourcePos, sourceExists := tablePositions[fk.ReferencedTable]
			targetPos, targetExists := tablePositions[table.Name]

			crossingDistance := 0
			if sourceExists && targetExists {
				crossingDistance = abs(sourcePos - targetPos)
			}

			relationships = append(relationships, Relationship{
				SourceTable:      fk.ReferencedTable,
				TargetTable:      table.Name,
				RelationName:     fk.RelationName,
				CrossingDistance: crossingDistance,
			})
		}
	}

	// Sort relationships to minimize crossings
	// Strategy: Draw shorter edges first (less likely to cross other edges)
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].CrossingDistance < relationships[j].CrossingDistance
	})

	// Process relationships in optimized order
	for _, rel := range relationships {
		relationship := fmt.Sprintf("    %s ||--o{ %s : \"%s\"\n",
			rel.SourceTable,
			rel.TargetTable,
			rel.RelationName)
		builder.WriteString(relationship)
	}

	return builder.String(), nil
}

// Helper function to calculate absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
