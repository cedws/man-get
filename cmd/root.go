package cmd

import (
	"fmt"
	"os"

	"github.com/cedws/man1c/man"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "debman",
	Short: "debman is a tool to search for manpages in Debian packages",
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
