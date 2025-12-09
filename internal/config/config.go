package config

// Config holds application configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Tables   string
	Format   string
}

// GetTablesList returns a slice of table names from the comma-separated list
func (c *Config) GetTablesList() []string {
	if c.Tables == "" {
		return nil
	}

	var tables []string
	currentTable := ""

	for _, char := range c.Tables {
		if char == ',' {
			if currentTable != "" {
				tables = append(tables, currentTable)
				currentTable = ""
			}
		} else if char != ' ' { // Ignore spaces
			currentTable += string(char)
		}
	}

	if currentTable != "" {
		tables = append(tables, currentTable)
	}

	return tables
}
