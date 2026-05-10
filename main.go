// Package main is the entry point for the GitHub MCP Server.
// It initializes and starts the Model Context Protocol server that
// provides GitHub API capabilities to AI assistants.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
	// Commit is set at build time via ldflags
	Commit = "none"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		token   string
		host    string
		logFile string
		readOnly bool
	)

	cmd := &cobra.Command{
		Use:   "github-mcp-server",
		Short: "GitHub MCP Server - Model Context Protocol server for GitHub",
		Long: `A Model Context Protocol (MCP) server that exposes GitHub API
capabilities to AI assistants. Supports repositories, issues, pull
requests, code search, and more.`,
		Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), token, host, logFile, readOnly)
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "GitHub personal access token (or set GITHUB_TOKEN env var)")
	cmd.Flags().StringVar(&host, "host", "https://api.github.com", "GitHub API host URL")
	cmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (defaults to stderr)")
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Run in read-only mode (disables write operations)")

	return cmd
}

func runServer(ctx context.Context, token, host, logFile string, readOnly bool) error {
	// Fall back to environment variable for token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if token == "" {
		return fmt.Errorf("GitHub token is required: set --token flag or GITHUB_TOKEN environment variable")
	}

	// Set up log output
	logOutput := os.Stderr
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer f.Close()
		logOutput = f
	}

	// Build server configuration
	cfg := server.Config{
		Token:    token,
		Host:     host,
		ReadOnly: readOnly,
		LogOutput: logOutput,
	}

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Handle OS signals for graceful shutdown
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Fprintf(logOutput, "Starting GitHub MCP Server %s\n", Version)

	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
