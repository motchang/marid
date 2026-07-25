package utils

import (
	"strings"
)

// SanitizeIdentifier sanitizes and escapes identifiers for diagram output
func SanitizeIdentifier(identifier string) string {
	// Replace spaces and special characters
	sanitized := strings.ReplaceAll(identifier, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")

	return sanitized
}

// columnTypeDisplayNames maps common MySQL types to more readable formats.
var columnTypeDisplayNames = map[string]string{
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

// FormatColumnType formats a column type for display
func FormatColumnType(dataType string) string {
	if display, ok := columnTypeDisplayNames[strings.ToLower(dataType)]; ok {
		return display
	}
	return dataType
}
