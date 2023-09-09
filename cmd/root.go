package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/cedws/man-get/internal/man"
	"github.com/spf13/cobra"
)

var sectionPattern = regexp.MustCompile("^[0-9]$")

var rootCmd = &cobra.Command{
	Use:   "man-get",
	Short: "CLI tool to grab Debian manpages",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pages := args[0:]
		sections := man.DefaultSections()

		if len(args) >= 2 && sectionPattern.MatchString(args[0]) {
			pages = args[1:]
			sections = []string{args[0]}
		}

		if err := man.Fetch(sections, pages); err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
