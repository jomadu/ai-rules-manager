package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arm",
	Short: "AI Rules Manager - A package manager for AI coding assistant rulesets",
	Long:  "ARM helps you install, update, and manage coding rules across different AI tools like Cursor and Amazon Q Developer.",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
