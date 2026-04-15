package main

import (
	"fmt"
	"os"

	"github.com/jgabor/gitfetch/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitfetch",
	Short: "Repo decay tracker — neofetch for git repos",
	Long:  "gitfetch scans git repos for staleness and displays a color-coded decay dashboard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, path, err := config.LoadOrCreate()
		if err != nil {
			return err
		}
		fmt.Printf("config: %s\n", path)
		fmt.Printf("repos: %v\n", cfg.Repos)
		return nil
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Scan repos and update cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := config.LoadOrCreate()
		if err != nil {
			return err
		}
		fmt.Printf("refreshing %d repos...\n", len(cfg.Repos))
		return nil
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive repo management",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := config.LoadOrCreate()
		if err != nil {
			return err
		}
		fmt.Printf("tui: %d repos configured\n", len(cfg.Repos))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(tuiCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
