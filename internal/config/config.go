package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Port         string   `json:"port"`
	BackendList  []string `json:"backend_list"`
	HealthCheck  bool     `json:"health_check"`
	HealthPeriod int      `json:"health_period"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:         "3000",
		BackendList:  []string{"http://localhost:8081", "http://localhost:8082"},
		HealthCheck:  true,
		HealthPeriod: 20,
	}
}

func EnsureConfigFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil // File already exists
	}

	if !os.IsNotExist(err) {
		return err // Some other OS error occurred
	}

	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Serialize default configuration
	config := DefaultConfig()
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	data = append(data, '\n') // Add a newline at the end for better readability

	// Write the file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	fmt.Printf("Created default configuration file at: %s\n", path)
	return nil
}

func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if len(c.BackendList) == 0 {
		return fmt.Errorf("backend list cannot be empty")
	}
	if c.HealthPeriod <= 0 {
		return fmt.Errorf("health check period must be greater than zero")
	}
	return nil
}

func (c *Config) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}
	return c.Validate()
}

func GetSystemConfigPath() string {
	var dir string

	switch runtime.GOOS {
		case "windows":
			// Resolves to C:\ProgramData
			dir = os.Getenv("ProgramData")
			if dir == "" {
				dir = `C:\ProgramData`
			}
		case "darwin":
			dir = "/Library/Preferences"
		default: // Linux / FreeBS/etc.
			dir = "/etc"
	}

	return filepath.Join(dir, "triv_lb", "config.json")
}