package core

import "github.com/spf13/cobra"

var versionsCmd = &cobra.Command{
	Use: "versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
