package cmd

import (
	"fmt"
	"os"

	"github.com/kanekitakitos/cmd-vault/internal/db"
	"github.com/kanekitakitos/cmd-vault/internal/tui"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	gitCommit = "none"
)

var rootCmd = &cobra.Command{
	Use:     "cmd-vault",
	Short:   "Cmd-Vault - retro TUI for saved shell commands",
	Version: fmt.Sprintf("%s (commit: %s)", version, gitCommit),
	Run: func(cmd *cobra.Command, args []string) {
		// When no args, start interactive TUI
		// Open DB
		store, err := db.Open(dbPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to open database:", err)
			os.Exit(1)
		}
		defer store.Close()

		if err := tui.RunTUI(store); err != nil {
			fmt.Fprintln(os.Stderr, "TUI error:", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error executing command:", err)
		os.Exit(1)
	}
}
