package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of telegraphcl",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("telegraphcl version", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
