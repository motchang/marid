package config

import "testing"

func TestGetTablesList(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "single table without spaces",
			input:    "users",
			expected: []string{"users"},
		},
		{
			name:     "multiple tables with spaces",
			input:    "users, orders , products",
			expected: []string{"users", "orders", "products"},
		},
		{
			name:     "trailing comma is ignored",
			input:    "users,orders,",
			expected: []string{"users", "orders"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Tables: tc.input}
			got := cfg.GetTablesList()

			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d tables, got %d", len(tc.expected), len(got))
			}

			for i, table := range tc.expected {
				if got[i] != table {
					t.Fatalf("index %d: expected %q, got %q", i, table, got[i])
				}
			}
		})
	}
}

func TestMergeWithCommandLineConfig(t *testing.T) {
	mycnf := &MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "mycnf-user",
		Password: "mycnf-pass",
		Database: "mycnf-db",
	}

	cmdCfg := &Config{
		Host:     "cli-host",
		Port:     3307,
		User:     "cli-user",
		Password: "cli-pass",
		Database: "cli-db",
		Tables:   "users,orders",
	}

	merged := MergeWithCommandLineConfig(mycnf, cmdCfg)

	if merged.Host != cmdCfg.Host {
		t.Fatalf("expected host %q, got %q", cmdCfg.Host, merged.Host)
	}

	if merged.Port != cmdCfg.Port {
		t.Fatalf("expected port %d, got %d", cmdCfg.Port, merged.Port)
	}

	if merged.User != cmdCfg.User {
		t.Fatalf("expected user %q, got %q", cmdCfg.User, merged.User)
	}

	if merged.Password != cmdCfg.Password {
		t.Fatalf("expected password %q, got %q", cmdCfg.Password, merged.Password)
	}

	if merged.Database != cmdCfg.Database {
		t.Fatalf("expected database %q, got %q", cmdCfg.Database, merged.Database)
	}

	if merged.Tables != cmdCfg.Tables {
		t.Fatalf("expected tables %q, got %q", cmdCfg.Tables, merged.Tables)
	}

	t.Run("falls back to mycnf values when cli is empty", func(t *testing.T) {
		cmdOnlyTables := &Config{Tables: "one"}
		mergedFallback := MergeWithCommandLineConfig(mycnf, cmdOnlyTables)

		if mergedFallback.Host != mycnf.Host {
			t.Fatalf("expected fallback host %q, got %q", mycnf.Host, mergedFallback.Host)
		}

		if mergedFallback.Port != mycnf.Port {
			t.Fatalf("expected fallback port %d, got %d", mycnf.Port, mergedFallback.Port)
		}

		if mergedFallback.Tables != cmdOnlyTables.Tables {
			t.Fatalf("expected tables %q, got %q", cmdOnlyTables.Tables, mergedFallback.Tables)
		}
	})
}
