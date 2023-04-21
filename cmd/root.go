package cmd

import (
	"fmt"
	"os"

	"github.com/cedws/man-get/man"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "man-get",
	Short: "Cross-platform CLI tool to grab Debian manpages",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		man.GetPages(args[0], args[1:])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
