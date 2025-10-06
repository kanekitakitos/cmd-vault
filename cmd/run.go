package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/kanekitakitos/cmd-vault/internal/db"
)

var dbPath string

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&dbPath, "db", "vault.db", "path to sqlite database file")
}

var runCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Run a saved command by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer store.Close()

		c, err := store.GetByName(name)
		if err != nil {
			return err
		}
		if c == nil {
			return fmt.Errorf("no command found with name %s", name)
		}

		// Use Windows cmd /C as specified
		execCmd := exec.Command("cmd", "/C", c.CommandStr)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}

		// increment usage count
		if err := store.IncrementUsage(c.ID); err != nil {
			return err
		}
		fmt.Println("Done.")
		return nil
	},
}
