package main

import (
	"fmt"
	"os"

	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	"github.com/jgabor/gitfetch/internal/display"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitfetch",
	Short: "Repo decay tracker — neofetch for git repos",
	Long:  "gitfetch scans git repos for staleness and displays a color-coded decay dashboard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, err := config.LoadOrCreate()
		if err != nil {
			return err
		}

		cachePath, err := config.CachePath()
		if err != nil {
			return fmt.Errorf("resolving cache path: %w", err)
		}

		c, err := cache.Load(cachePath)
		if err != nil {
			return fmt.Errorf("loading cache: %w", err)
		}

		fmt.Print(display.FormatDashboard(c.Repos))
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

		results := gitscanner.ScanAll(cfg.Repos)

		cachePath, err := config.CachePath()
		if err != nil {
			return fmt.Errorf("resolving cache path: %w", err)
		}

		c, err := cache.Load(cachePath)
		if err != nil {
			return fmt.Errorf("loading cache: %w", err)
		}

		ok, fail := 0, 0
		for _, r := range results {
			c.Repos[r.RepoPath] = cache.RepoEntry{
				LastCommitDate: r.LastCommitDate,
				LastTagDate:    r.LastTagDate,
				Error:          r.Error,
				ScannedAt:      r.ScannedAt,
			}
			if r.Error != "" {
				fmt.Fprintf(os.Stderr, "  error: %s: %s\n", r.RepoPath, r.Error)
				fail++
			} else {
				fmt.Printf("  ok: %s\n", r.RepoPath)
				ok++
			}
		}

		if err := cache.Save(cachePath, c); err != nil {
			return fmt.Errorf("saving cache: %w", err)
		}

		fmt.Printf("\nrefreshed %d repos: %d ok, %d failed\n", len(results), ok, fail)
		fmt.Print(display.FormatDashboard(c.Repos))
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
