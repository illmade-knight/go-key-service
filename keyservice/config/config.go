package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config defines the full configuration for the Key Service.
type Config struct {
	RunMode            string `yaml:"run_mode"`
	ProjectID          string `yaml:"project_id"`
	HTTPListenAddr     string `yaml:"http_listen_addr"`
	IdentityServiceURL string `yaml:"identity_service_url"`

	// CORS configuration is now also loaded from the YAML file.
	Cors struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"cors"`
}

// Load reads a YAML file from the given path and returns a Config struct.
// It will override fields with environment variables where appropriate.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &cfg, nil
}
