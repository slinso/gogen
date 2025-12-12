package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration.
type Config struct {
	TypeMappings map[string]string `yaml:"typeMappings" json:"typeMappings"`
	Options      Options           `yaml:"options" json:"options"`
}

// Options represents generation options.
type Options struct {
	PerType      bool     `yaml:"perType" json:"perType"`
	ExportedOnly bool     `yaml:"exportedOnly" json:"exportedOnly"`
	TagKey       string   `yaml:"tagKey" json:"tagKey"`
	IncludeTypes []string `yaml:"includeTypes" json:"includeTypes"`
	ExcludeTypes []string `yaml:"excludeTypes" json:"excludeTypes"`
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		TypeMappings: DefaultTypeMappings(),
		Options:      DefaultOptions(),
	}
}

// LoadFile loads configuration from a file (YAML or JSON based on extension).
func (c *Config) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	var loaded Config
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &loaded); err != nil {
			return fmt.Errorf("parsing YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &loaded); err != nil {
			return fmt.Errorf("parsing JSON config: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &loaded); err != nil {
			if err := json.Unmarshal(data, &loaded); err != nil {
				return fmt.Errorf("unable to parse config as YAML or JSON")
			}
		}
	}

	// Merge loaded config with defaults
	c.merge(&loaded)

	return nil
}

// merge merges the loaded config into the current config.
func (c *Config) merge(loaded *Config) {
	// Merge type mappings (loaded values override defaults)
	if loaded.TypeMappings != nil {
		for k, v := range loaded.TypeMappings {
			c.TypeMappings[k] = v
		}
	}

	// Merge options
	if loaded.Options.TagKey != "" {
		c.Options.TagKey = loaded.Options.TagKey
	}
	if loaded.Options.PerType {
		c.Options.PerType = true
	}
	// ExportedOnly defaults to true, so we check if it was explicitly set to false
	c.Options.ExportedOnly = loaded.Options.ExportedOnly
	c.Options.IncludeTypes = loaded.Options.IncludeTypes
	c.Options.ExcludeTypes = loaded.Options.ExcludeTypes
}

// MapType maps a Go type to its target type using the configured mappings.
func (c *Config) MapType(goType string) string {
	if mapped, ok := c.TypeMappings[goType]; ok {
		return mapped
	}
	return goType
}

// ShouldIncludeType checks if a type should be included based on config.
func (c *Config) ShouldIncludeType(name string, isExported bool) bool {
	// Check exported only filter
	if c.Options.ExportedOnly && !isExported {
		return false
	}

	// Check include list (if specified, type must be in it)
	if len(c.Options.IncludeTypes) > 0 {
		found := false
		for _, t := range c.Options.IncludeTypes {
			if t == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude list
	for _, t := range c.Options.ExcludeTypes {
		if t == name {
			return false
		}
	}

	return true
}
