package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jomadu/arm/internal/config"
	"gopkg.in/ini.v1"
)

func configCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config command requires a subcommand: list, get, or set")
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		return configList()
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("config get requires a key")
		}
		return configGet(args[1])
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("config set requires a key and value")
		}
		return configSet(args[1], args[2])
	default:
		return fmt.Errorf("unknown config subcommand: %s", subcommand)
	}
}

func configList() error {
	manager := config.NewManager()
	if err := manager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	fmt.Println("Configuration:")
	fmt.Println()

	// List sources
	if len(cfg.Sources) > 0 {
		fmt.Println("[sources]")
		for name, source := range cfg.Sources {
			fmt.Printf("%s = %s\n", name, source.URL)
		}
		fmt.Println()

		// List source-specific configs
		for name, source := range cfg.Sources {
			if source.AuthToken != "" || source.Timeout != "" {
				fmt.Printf("[sources.%s]\n", name)
				if source.AuthToken != "" {
					// Mask auth tokens for security
					maskedToken := maskToken(source.AuthToken)
					fmt.Printf("authToken = %s\n", maskedToken)
				}
				if source.Timeout != "" {
					fmt.Printf("timeout = %s\n", source.Timeout)
				}
				fmt.Println()
			}
		}
	}

	// List cache config
	if cfg.Cache.Location != "" || cfg.Cache.MaxSize != "" {
		fmt.Println("[cache]")
		if cfg.Cache.Location != "" {
			fmt.Printf("location = %s\n", cfg.Cache.Location)
		}
		if cfg.Cache.MaxSize != "" {
			fmt.Printf("maxSize = %s\n", cfg.Cache.MaxSize)
		}
	}

	return nil
}

func configGet(key string) error {
	manager := config.NewManager()
	if err := manager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	value, err := getConfigValue(cfg, key)
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}

func configSet(key, value string) error {
	// Determine which config file to write to (prefer project-level)
	configPath := ".armrc"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If no project config exists, use user config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".armrc")
	}

	// Load existing config or create new one
	var cfg *ini.File
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg = ini.Empty()
	} else {
		var err error
		cfg, err = ini.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Set the value
	if err := setConfigValue(cfg, key, value); err != nil {
		return err
	}

	// Save the config file
	if err := cfg.SaveTo(configPath); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func getConfigValue(cfg *config.ARMConfig, key string) (string, error) {
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "sources":
		if len(parts) == 2 {
			// sources.name -> URL
			if source, exists := cfg.Sources[parts[1]]; exists {
				return source.URL, nil
			}
			return "", fmt.Errorf("source %s not found", parts[1])
		} else if len(parts) == 3 {
			// sources.name.field
			sourceName := parts[1]
			field := parts[2]
			if source, exists := cfg.Sources[sourceName]; exists {
				switch field {
				case "authToken":
					return maskToken(source.AuthToken), nil
				case "timeout":
					return source.Timeout, nil
				default:
					return "", fmt.Errorf("unknown source field: %s", field)
				}
			}
			return "", fmt.Errorf("source %s not found", sourceName)
		}
	case "cache":
		if len(parts) == 2 {
			switch parts[1] {
			case "location":
				return cfg.Cache.Location, nil
			case "maxSize":
				return cfg.Cache.MaxSize, nil
			default:
				return "", fmt.Errorf("unknown cache field: %s", parts[1])
			}
		}
	}

	return "", fmt.Errorf("unknown config key: %s", key)
}

func setConfigValue(cfg *ini.File, key, value string) error {
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "sources":
		if len(parts) == 2 {
			// sources.name -> URL
			section := cfg.Section("sources")
			section.Key(parts[1]).SetValue(value)
			return nil
		} else if len(parts) == 3 {
			// sources.name.field
			sourceName := parts[1]
			field := parts[2]
			sectionName := fmt.Sprintf("sources.%s", sourceName)
			section := cfg.Section(sectionName)
			section.Key(field).SetValue(value)
			return nil
		}
	case "cache":
		if len(parts) == 2 {
			section := cfg.Section("cache")
			section.Key(parts[1]).SetValue(value)
			return nil
		}
	}

	return fmt.Errorf("unknown config key: %s", key)
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
