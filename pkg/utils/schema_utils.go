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

// FormatColumnType formats a column type for display
func FormatColumnType(dataType string) string {
	// Map common MySQL types to more readable formats
	switch strings.ToLower(dataType) {
	case "int", "tinyint", "smallint", "mediumint", "bigint":
		return "integer"
	case "float", "double", "decimal":
		return "float"
	case "varchar", "char", "text", "tinytext", "mediumtext", "longtext":
		return "string"
	case "datetime", "timestamp":
		return "datetime"
	case "date":
		return "date"
	case "time":
		return "time"
	case "blob", "tinyblob", "mediumblob", "longblob":
		return "blob"
	case "boolean", "bool":
		return "boolean"
	case "enum", "set":
		return "enum"
	case "json":
		return "json"
	default:
		return dataType
	}
}
