package utils

import "testing"

func TestSanitizeIdentifier(t *testing.T) {
	cases := map[string]string{
		"plain":               "plain",
		"with spaces":         "with_spaces",
		"with-dash-and space": "with_dash_and_space",
	}

	for input, expected := range cases {
		got := SanitizeIdentifier(input)
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
}

func TestFormatColumnType(t *testing.T) {
	cases := map[string]string{
		"int":       "integer",
		"SMALLINT":  "integer",
		"double":    "float",
		"varchar":   "string",
		"timestamp": "datetime",
		"date":      "date",
		"json":      "json",
		"custom":    "custom",
	}

	for input, expected := range cases {
		got := FormatColumnType(input)
		if got != expected {
			t.Fatalf("%s: expected %q, got %q", input, expected, got)
		}
	}
}
