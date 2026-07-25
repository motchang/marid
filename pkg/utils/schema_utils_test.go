package utils

import "testing"

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain identifier is unchanged", input: "plain", want: "plain"},
		{name: "space becomes underscore", input: "with spaces", want: "with_spaces"},
		{name: "dash becomes underscore", input: "with-dash", want: "with_dash"},
		{name: "dashes and spaces mixed", input: "with-dash-and space", want: "with_dash_and_space"},
		{name: "each repeated separator becomes its own underscore", input: "a  b--c", want: "a__b__c"},
		{name: "leading and trailing separators are kept as underscores", input: " lead-trail ", want: "_lead_trail_"},
		{name: "existing underscores are left alone", input: "already_ok", want: "already_ok"},
		{name: "other punctuation is not sanitized", input: "keeps.dot$sign", want: "keeps.dot$sign"},
		{name: "empty string", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeIdentifier(tt.input); got != tt.want {
				t.Errorf("SanitizeIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// formatColumnTypeCases must hold one case per entry in
// columnTypeDisplayNames. TestFormatColumnTypeCoversEveryMapping enforces
// that, so a new mapping cannot be added without asserting what it renders as.
var formatColumnTypeCases = map[string]string{
	"int":        "integer",
	"tinyint":    "integer",
	"smallint":   "integer",
	"mediumint":  "integer",
	"bigint":     "integer",
	"float":      "float",
	"double":     "float",
	"decimal":    "float",
	"varchar":    "string",
	"char":       "string",
	"text":       "string",
	"tinytext":   "string",
	"mediumtext": "string",
	"longtext":   "string",
	"datetime":   "datetime",
	"timestamp":  "datetime",
	"date":       "date",
	"time":       "time",
	"blob":       "blob",
	"tinyblob":   "blob",
	"mediumblob": "blob",
	"longblob":   "blob",
	"boolean":    "boolean",
	"bool":       "boolean",
	"enum":       "enum",
	"set":        "enum",
	"json":       "json",
}

func TestFormatColumnType(t *testing.T) {
	for input, want := range formatColumnTypeCases {
		t.Run(input, func(t *testing.T) {
			if got := FormatColumnType(input); got != want {
				t.Errorf("FormatColumnType(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

// TestFormatColumnTypeCoversEveryMapping guards formatColumnTypeCases against
// drifting out of sync with the lookup table it is meant to pin down.
func TestFormatColumnTypeCoversEveryMapping(t *testing.T) {
	for dataType := range columnTypeDisplayNames {
		if _, ok := formatColumnTypeCases[dataType]; !ok {
			t.Errorf("columnTypeDisplayNames has %q but formatColumnTypeCases does not assert its display name", dataType)
		}
	}

	for dataType := range formatColumnTypeCases {
		if _, ok := columnTypeDisplayNames[dataType]; !ok {
			t.Errorf("formatColumnTypeCases asserts %q but columnTypeDisplayNames no longer maps it", dataType)
		}
	}
}

func TestFormatColumnTypeIsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "all caps", input: "SMALLINT", want: "integer"},
		{name: "all caps with prefix", input: "TINYBLOB", want: "blob"},
		{name: "short alias in caps", input: "BOOL", want: "boolean"},
		{name: "mixed case", input: "VarChar", want: "string"},
		{name: "title case", input: "Json", want: "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatColumnType(tt.input); got != tt.want {
				t.Errorf("FormatColumnType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestFormatColumnTypeFallsBackToInput pins the lookup miss: unmapped types are
// echoed back verbatim, including the original casing. Note that only bare type
// names are mapped, so a parameterized or qualified MySQL type passes through.
func TestFormatColumnTypeFallsBackToInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unmapped type", input: "geometry"},
		{name: "arbitrary string", input: "custom"},
		{name: "parameterized type keeps its length", input: "varchar(255)"},
		{name: "type with an attribute", input: "int unsigned"},
		{name: "casing is preserved on a miss", input: "GeoMetry"},
		{name: "empty string", input: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatColumnType(tt.input); got != tt.input {
				t.Errorf("FormatColumnType(%q) = %q, want the input echoed back", tt.input, got)
			}
		})
	}
}
