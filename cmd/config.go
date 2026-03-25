package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Config stores CLI configuration.
type Config struct {
	Token  string `json:"token,omitempty"`
	APIURL string `json:"api_url,omitempty"`
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  "Set a configuration value. Supported keys: token, api-url",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := loadConfig()
		if err != nil {
			cfg = &Config{}
		}

		switch key {
		case "token":
			cfg.Token = value
		case "api-url":
			cfg.APIURL = value
		default:
			return fmt.Errorf("unknown config key: %s (supported: token, api-url)", key)
		}

		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Set %s\n", key)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		switch args[0] {
		case "token":
			if cfg.Token == "" {
				fmt.Println("(not set)")
			} else {
				// Show only prefix for security
				if len(cfg.Token) > 8 {
					fmt.Printf("%s...%s\n", cfg.Token[:4], cfg.Token[len(cfg.Token)-4:])
				} else {
					fmt.Println("(set)")
				}
			}
		case "api-url":
			if cfg.APIURL == "" {
				fmt.Println(getDefaultAPIURL())
			} else {
				fmt.Println(cfg.APIURL)
			}
		default:
			return fmt.Errorf("unknown config key: %s", args[0])
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pipeboard"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")
	return os.WriteFile(path, data, 0600)
}
