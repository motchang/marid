package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/motchang/marid/internal/config"
	"github.com/motchang/marid/internal/database"
	"github.com/motchang/marid/internal/diagram"
	"github.com/motchang/marid/internal/schema"
	"github.com/motchang/marid/pkg/formatter"
	"github.com/spf13/cobra"
)

var (
	cfgHost       string
	cfgPort       int
	cfgUser       string
	cfgPassword   string
	cfgDatabase   string
	cfgTables     string
	cfgFormat     string
	cfgPromptPass bool
	cfgUseMyCnf   bool
	cfgNoPassword bool

	getMyCnfConfig    = config.GetMyCnfConfig
	promptForPassword = config.PromptForPassword
	connect           = database.Connect
	extract           = schema.Extract
	generate          = diagram.Generate
)

func main() {
	rootCmd := buildRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "marid",
		Short: "MySQL to Mermaid ER Diagram Generator",
		Long: `Marid connects to a MySQL database, extracts table definitions,
and generates Mermaid ER diagrams based on the schema.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdConfig := config.Config{
				Host:     cfgHost,
				Port:     cfgPort,
				User:     cfgUser,
				Password: cfgPassword,
				Database: cfgDatabase,
				Tables:   cfgTables,
				Format:   cfgFormat,
			}

			cfg, err := resolveConfig(cmd, cmdConfig)
			if err != nil {
				return err
			}

			db, err := connect(cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}

			if db != nil {
				defer func(db *sql.DB) {
					_ = db.Close()
				}(db)
			}

			dbSchema, err := extract(db, cfg)
			if err != nil {
				return fmt.Errorf("failed to extract schema: %w", err)
			}

			mermaidDiagram, err := generate(dbSchema, cfg.Format)
			if err != nil {
				return fmt.Errorf("failed to generate diagram: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), mermaidDiagram)
			return err
		},
	}

	// Use shorthand-enabled flag helpers (VarP/VarP) to match the documented short options.
	availableFormats := formatter.Available()
	formatDesc := fmt.Sprintf("Output format (default: %s)", formatter.DefaultFormat)
	if len(availableFormats) > 0 {
		formatDesc += fmt.Sprintf("; available: %s", strings.Join(availableFormats, ", "))
	}

	rootCmd.Flags().StringVarP(&cfgHost, "host", "H", "localhost", "MySQL host address")
	rootCmd.Flags().IntVarP(&cfgPort, "port", "P", 3306, "MySQL port")
	rootCmd.Flags().StringVarP(&cfgUser, "user", "u", "root", "MySQL username")
	rootCmd.Flags().StringVarP(&cfgPassword, "password", "p", "", "MySQL password (insecure, prefer --ask-password)")
	rootCmd.Flags().BoolVar(&cfgPromptPass, "ask-password", false, "Prompt for password (secure)")
	rootCmd.Flags().BoolVarP(&cfgUseMyCnf, "use-mycnf", "c", false, "Read connection info from ~/.my.cnf")
	rootCmd.Flags().BoolVarP(&cfgNoPassword, "no-password", "n", false, "Connect without a password")
	rootCmd.Flags().StringVarP(&cfgDatabase, "database", "d", "", "Database name (required)")
	rootCmd.Flags().StringVarP(&cfgTables, "tables", "t", "", "Comma-separated list of tables (default: all tables)")
	rootCmd.Flags().StringVarP(&cfgFormat, "format", "f", formatter.DefaultFormat, formatDesc)

	return rootCmd
}

// passwordFlagConflict lists the password flags the user selected when more than
// one of them is in play, and nil otherwise. The three name contradictory
// sources for one value, so combining them is rejected rather than resolved by a
// silent precedence.
//
// Booleans are judged by value, not by whether they were provided: cobra's
// MarkFlagsMutuallyExclusive would do the latter, and pflag records
// --no-password=false as set, so an explicitly disabled flag would count as a
// conflict. --use-mycnf is deliberately not a member — a password read from
// ~/.my.cnf combined with --ask-password or --no-password is an override rather
// than a contradiction.
func passwordFlagConflict(cmd *cobra.Command) []string {
	var selected []string

	if cmd.Flags().Changed("password") {
		selected = append(selected, "--password")
	}

	if cfgNoPassword {
		selected = append(selected, "--no-password")
	}

	if cfgPromptPass {
		selected = append(selected, "--ask-password")
	}

	if len(selected) < 2 {
		return nil
	}

	return selected
}

// resolveConfig merges ~/.my.cnf settings (when requested) into cmdConfig,
// validates the result, and applies password overrides.
func resolveConfig(cmd *cobra.Command, cmdConfig config.Config) (config.Config, error) {
	cfg := cmdConfig

	if conflicting := passwordFlagConflict(cmd); conflicting != nil {
		return cfg, fmt.Errorf("conflicting password flags: %s; specify only one",
			strings.Join(conflicting, ", "))
	}

	if cfgUseMyCnf {
		myCnfConfig, err := getMyCnfConfig()
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Could not read .my.cnf: %v\n", err)
		} else {
			cfg = *config.MergeWithCommandLineConfig(myCnfConfig, &cmdConfig)
		}
	}

	if cfg.Database == "" {
		return cfg, fmt.Errorf("database name is required")
	}

	if cfgNoPassword {
		cfg.Password = ""
	} else if cfgPromptPass {
		password, err := promptForPassword()
		if err != nil {
			return cfg, fmt.Errorf("failed to read password: %w", err)
		}
		cfg.Password = password
	}

	return cfg, nil
}
