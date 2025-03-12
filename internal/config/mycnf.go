package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
)

// MySQLConfig represents MySQL connection configuration from my.cnf
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// GetMyCnfConfig reads MySQL configuration from .my.cnf files
func GetMyCnfConfig() (*MySQLConfig, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Path to .my.cnf file
	myCnfPath := filepath.Join(homeDir, ".my.cnf")

	// Check if file exists
	if _, err := os.Stat(myCnfPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".my.cnf file not found at %s", myCnfPath)
	}

	// Load the INI file
	cfg, err := ini.Load(myCnfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .my.cnf file: %w", err)
	}

	// Try to read from [client] section first
	clientSection := cfg.Section("client")

	// Initialize with default values
	config := &MySQLConfig{
		Host: "localhost",
		Port: 3306,
	}

	// Update with values from config file if they exist
	if clientSection.HasKey("host") {
		config.Host = clientSection.Key("host").String()
	}

	if clientSection.HasKey("port") {
		config.Port, _ = clientSection.Key("port").Int()
	}

	if clientSection.HasKey("user") {
		config.User = clientSection.Key("user").String()
	}

	if clientSection.HasKey("password") {
		config.Password = clientSection.Key("password").String()
	}

	if clientSection.HasKey("database") {
		config.Database = clientSection.Key("database").String()
	}

	return config, nil
}

// MergeWithCommandLineConfig merges my.cnf values with command line values
// Command line values take precedence over my.cnf values
func MergeWithCommandLineConfig(myCnfConfig *MySQLConfig, cmdConfig *Config) *Config {
	// Start with my.cnf values
	mergedConfig := &Config{
		Host:     myCnfConfig.Host,
		Port:     myCnfConfig.Port,
		User:     myCnfConfig.User,
		Password: myCnfConfig.Password,
		Database: myCnfConfig.Database,
		Tables:   cmdConfig.Tables, // Tables are only specified via command line
	}

	// Override with command line values if they're not empty
	if cmdConfig.Host != "" {
		mergedConfig.Host = cmdConfig.Host
	}

	if cmdConfig.Port != 0 {
		mergedConfig.Port = cmdConfig.Port
	}

	if cmdConfig.User != "" {
		mergedConfig.User = cmdConfig.User
	}

	if cmdConfig.Password != "" {
		mergedConfig.Password = cmdConfig.Password
	}

	if cmdConfig.Database != "" {
		mergedConfig.Database = cmdConfig.Database
	}

	return mergedConfig
}
