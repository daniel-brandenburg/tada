package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultSort   string   `yaml:"default_sort"`
	Theme         string   `yaml:"theme"`
	DefaultStatus string   `yaml:"default_status"`
	Tags          []string `yaml:"tags"`
}

func getConfigPaths() (global, local string) {
	home, _ := os.UserHomeDir()
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".config")
	}
	global = filepath.Join(xdg, "tada", "config.yaml")
	local = filepath.Join(".tada", "config.yaml")
	return
}

func loadConfig() (*Config, error) {
	globalPath, localPath := getConfigPaths()
	cfg := &Config{}
	// Load global config
	if data, err := os.ReadFile(globalPath); err == nil {
		yaml.Unmarshal(data, cfg)
	}
	// Override with local config
	if data, err := os.ReadFile(localPath); err == nil {
		yaml.Unmarshal(data, cfg)
	}
	return cfg, nil
}

func saveConfig(cfg *Config, global bool) error {
	globalPath, localPath := getConfigPaths()
	path := localPath
	if global {
		path = globalPath
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := yaml.Marshal(cfg)
	return os.WriteFile(path, data, 0644)
}

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or set configuration",
		Long:  "View or set global/local configuration for tada.",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show effective config",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := loadConfig()
			data, _ := yaml.Marshal(cfg)
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value] [--global]",
		Short: "Set a config value",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key, value := args[0], args[1]
			global, _ := cmd.Flags().GetBool("global")
			cfg, _ := loadConfig()
			switch key {
			case "default_sort":
				cfg.DefaultSort = value
			case "theme":
				cfg.Theme = value
			case "default_status":
				cfg.DefaultStatus = value
			case "tags":
				cfg.Tags = strings.Split(value, ",")
			default:
				fmt.Fprintln(cmd.ErrOrStderr(), lipgloss.NewStyle().Foreground(cliError).Render("Unknown config key."))
				return
			}
			if err := saveConfig(cfg, global); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), lipgloss.NewStyle().Foreground(cliError).Render("Failed to save config."))
				return
			}
			fmt.Fprintln(cmd.OutOrStdout(), lipgloss.NewStyle().Foreground(cliPrimary).Bold(true).Render("Config updated."))
		},
	})
	cmd.PersistentFlags().Bool("global", false, "Affect global config instead of local")
	return cmd
}
