package main

import (
	"fmt"
	"os"

	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	"github.com/jgabor/gitfetch/internal/display"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
	"github.com/jgabor/gitfetch/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitfetch",
	Short: "Repo decay tracker — neofetch for git repos",
	Long:  "gitfetch scans git repos for staleness and displays a color-coded decay dashboard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := config.LoadOrCreate()
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

		fmt.Print(display.FormatDashboard(cache.FilterByRepos(c, cfg.Repos)))
		return nil
	},
}

var (
	authorFilter string
	remoteFilter string
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Scan repos and update cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := config.LoadOrCreate()
		if err != nil {
			return err
		}

		opts := gitscanner.ScanOptions{
			Author: authorFilter,
			Remote: remoteFilter,
		}
		results := gitscanner.ScanAll(cfg.Repos, opts)

		cachePath, err := config.CachePath()
		if err != nil {
			return fmt.Errorf("resolving cache path: %w", err)
		}

		c, err := cache.Load(cachePath)
		if err != nil {
			return fmt.Errorf("loading cache: %w", err)
		}

		repoSet := make(map[string]bool, len(cfg.Repos))
		for _, r := range cfg.Repos {
			repoSet[r] = true
		}
		for key := range c.Repos {
			if !repoSet[key] {
				delete(c.Repos, key)
			}
		}

		ok, fail := 0, 0
		for _, r := range results {
			c.Repos[r.RepoPath] = cache.RepoEntry{
				LastCommitDate: r.LastCommitDate,
				LastTagDate:    r.LastTagDate,
				LastTag:        r.LastTag,
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
		cfg, cfgPath, err := config.LoadOrCreate()
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

		return tui.Run(cfg, cfgPath, c, cachePath)
	},
}

func init() {
	refreshCmd.Flags().StringVar(&authorFilter, "author", "", "filter repos by author name (substring match)")
	refreshCmd.Flags().StringVar(&remoteFilter, "remote", "", "filter repos by remote URL (substring match)")
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(tuiCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
