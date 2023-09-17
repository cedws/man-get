package cmd

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/cedws/man-get/internal/deb"
	"github.com/cedws/man-get/internal/man"
	"github.com/spf13/cobra"
)

const examples = `man-get tar ed
man-get 1 haproxy`

var sectionPattern = regexp.MustCompile("^[0-9]$")

var (
	mirror  string
	release string
)

func init() {
	rootCmd.Flags().StringVarP(&mirror, "mirror", "m", "https://ftp.debian.org/debian", "Debian mirror to use")
	rootCmd.Flags().StringVarP(&release, "release", "r", "bookworm", "Debian release to use")
}

func cacheDir() (string, error) {
	if cacheHome, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		return cacheHome, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".cache", "man-get"), nil
}

var rootCmd = &cobra.Command{
	Use:     "man-get [section] <page>...",
	Short:   "CLI tool to grab Debian manpages",
	Example: examples,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pages := args[0:]
		sections := man.DefaultSections()

		if len(args) >= 2 && sectionPattern.MatchString(args[0]) {
			pages = args[1:]
			sections = []string{args[0]}
		}

		cacheDir, err := cacheDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}

		client := deb.NewAptClient(
			deb.WithMirror(mirror),
			deb.WithDistribution(release),
			deb.WithArch("amd64"),
			deb.WithCacheDir(cacheDir),
		)

		if err := man.Fetch(client, sections, pages); err != nil {
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
