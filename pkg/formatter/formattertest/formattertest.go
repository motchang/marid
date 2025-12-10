package formattertest

import (
	"errors"

	"github.com/motchang/marid/pkg/formatter"
)

// SampleRenderData returns a canonical render data set used in formatter contract tests.
func SampleRenderData() formatter.RenderData {
	return formatter.RenderData{
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
}

// SampleMermaidOutput returns the expected Mermaid output for SampleRenderData.
func SampleMermaidOutput() string {
	return `erDiagram
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
}

// MockFormatter is a configurable formatter implementation intended for tests.
type MockFormatter struct {
	NameValue      string
	MediaTypeValue string
	RenderFunc     func(formatter.RenderData) (string, error)
	RenderCalls    []formatter.RenderData
}

// Name returns the configured name or "mock" if unset.
func (m *MockFormatter) Name() string {
	if m.NameValue != "" {
		return m.NameValue
	}
	return "mock"
}

// MediaType returns the configured media type or "text/plain" if unset.
func (m *MockFormatter) MediaType() string {
	if m.MediaTypeValue != "" {
		return m.MediaTypeValue
	}
	return "text/plain"
}

// Render records the call and delegates to RenderFunc when provided.
func (m *MockFormatter) Render(data formatter.RenderData) (string, error) {
	m.RenderCalls = append(m.RenderCalls, data)
	if m.RenderFunc != nil {
		return m.RenderFunc(data)
	}
	return "", errors.New("RenderFunc not provided")
}
