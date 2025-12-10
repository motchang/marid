package formatter

// Formatter defines the contract for rendering database schemas into specific output formats.
type Formatter interface {
	// Name returns the formatter name (e.g., "mermaid").
	Name() string
	// MediaType returns the MIME type associated with the output (e.g., "text/plain").
	MediaType() string
	// Render builds the formatted representation for the provided render data.
	Render(RenderData) (string, error)
}

// Factory constructs a Formatter instance.
type Factory func() Formatter

// DefaultFormat is the fallback format name when none is provided.
const DefaultFormat = "mermaid"

// RenderData represents normalized schema information passed to formatters.
type RenderData struct {
	Tables []Table
}

// Table represents a database table for rendering purposes.
type Table struct {
	Name        string
	Comment     string
	Columns     []Column
	PrimaryKey  []string
	ForeignKeys []ForeignKey
}

// Column represents a database column for rendering purposes.
type Column struct {
	Name       string
	DataType   string
	IsNullable bool
	IsPrimary  bool
	IsUnique   bool
	Comment    string
}

// ForeignKey represents a foreign key relationship for rendering purposes.
type ForeignKey struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	RelationName     string
}
