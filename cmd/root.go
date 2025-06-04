package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "seasonctl",
	Short: "CLI tool to manage sport seasons",
	Long: `seasonctl is a command-line tool to help you create and manage
sports seasons and schedule weeks in a PostgreSQL database.`,
}

// Execute runs the root command and is called from main.go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
