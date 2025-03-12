package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/motchang/marid/internal/config"
	"github.com/motchang/marid/internal/database"
	"github.com/motchang/marid/internal/diagram"
	"github.com/motchang/marid/internal/schema"
	"github.com/spf13/cobra"
)

var (
	cfgHost       string
	cfgPort       int
	cfgUser       string
	cfgPassword   string
	cfgDatabase   string
	cfgTables     string
	cfgPromptPass bool
	cfgUseMyCnf   bool
	cfgNoPassword bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "marid",
		Short: "MySQL to Mermaid ER Diagram Generator",
		Long: `Marid connects to a MySQL database, extracts table definitions,
and generates Mermaid ER diagrams based on the schema.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the configuration with command line values
			cmdConfig := config.Config{
				Host:     cfgHost,
				Port:     cfgPort,
				User:     cfgUser,
				Password: cfgPassword,
				Database: cfgDatabase,
				Tables:   cfgTables,
			}

			var cfg config.Config

			// Process configuration sources in order of precedence:
			// 1. Command line args (already in cmdConfig)
			// 2. .my.cnf file (if --use-mycnf is specified)
			// 3. Password prompt (if --ask-password is specified)

			if cfgUseMyCnf {
				// Try to load .my.cnf configuration
				myCnfConfig, err := config.GetMyCnfConfig()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not read .my.cnf: %v\n", err)
				} else {
					// Merge .my.cnf values with command line values
					mergedConfig := config.MergeWithCommandLineConfig(myCnfConfig, &cmdConfig)
					cfg = *mergedConfig
				}
			} else {
				// Use command line config directly
				cfg = cmdConfig
			}

			// If no database is specified, it's required
			if cfg.Database == "" {
				return fmt.Errorf("database name is required")
			}

			// Handle password according to flags
			if cfgNoPassword {
				// Explicitly use no password
				cfg.Password = ""
			} else if cfgPromptPass {
				// Prompt for password (highest security, overrides other password sources)
				password, err := config.PromptForPassword()
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				cfg.Password = password
			}

			// Connect to database
			db, err := database.Connect(cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer func(db *sql.DB) {
				_ = db.Close()
			}(db)

			// Extract schema
			dbSchema, err := schema.Extract(db, cfg)
			if err != nil {
				return fmt.Errorf("failed to extract schema: %w", err)
			}

			// Generate diagram
			mermaidDiagram, err := diagram.Generate(dbSchema)
			if err != nil {
				return fmt.Errorf("failed to generate diagram: %w", err)
			}

			// Output diagram
			fmt.Println(mermaidDiagram)
			return nil
		},
	}

	// Define flags without shorthands
	rootCmd.Flags().StringVar(&cfgHost, "host", "localhost", "MySQL host address")
	rootCmd.Flags().IntVar(&cfgPort, "port", 3306, "MySQL port")
	rootCmd.Flags().StringVar(&cfgUser, "user", "root", "MySQL username")
	rootCmd.Flags().StringVar(&cfgPassword, "password", "", "MySQL password (insecure, prefer --ask-password)")
	rootCmd.Flags().BoolVar(&cfgPromptPass, "ask-password", false, "Prompt for password (secure)")
	rootCmd.Flags().BoolVar(&cfgUseMyCnf, "use-mycnf", false, "Read connection info from ~/.my.cnf")
	rootCmd.Flags().BoolVar(&cfgNoPassword, "no-password", false, "Connect without a password")
	rootCmd.Flags().StringVar(&cfgDatabase, "database", "", "Database name (required)")
	rootCmd.Flags().StringVar(&cfgTables, "tables", "", "Comma-separated list of tables (default: all tables)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
