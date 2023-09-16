package man

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cedws/man-get/internal/deb"
	"github.com/laher/argo/ar"
	"github.com/ulikunitz/xz"
)

const finishMessage = `Append %v to your MANPATH to open the downloaded manpages:

	Bash/Zsh:
	$ export MANPATH="$MANPATH:%v"

	Fish:
	$ set -x MANPATH "$MANPATH:%v"
`

// DefaultSections returns the default man page sections.
func DefaultSections() []string {
	return []string{
		"1", "n", "l", "8", "3", "0", "2", "3posix", "3pm", "3perl", "3am", "5", "4", "9", "6", "7",
	}
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

func dataHomeDir() (string, error) {
	if dataHome, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return dataHome, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".local/share", "man-get"), nil
}

func sectionPath(section string) (string, error) {
	dataHome, err := dataHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(dataHome, fmt.Sprintf("man%c", section[0])), nil
}

func pagePath(page, section string) (string, error) {
	sectionPath, err := sectionPath(section)
	if err != nil {
		return "", err
	}
	return path.Join(sectionPath, fmt.Sprintf("%v.%v.gz", page, section)), nil
}

func Fetch(desiredSections, desiredPages []string) error {
	cacheDir, err := cacheDir()
	if err != nil {
		return err
	}

	client := deb.NewAptClient(
		deb.WithMirror("https://ftp.debian.org/debian"),
		deb.WithDistribution("bullseye"),
		deb.WithArch("amd64"),
		deb.WithCacheDir(cacheDir),
	)

	for _, page := range desiredPages {
		sections, err := findSectionsForPage(client, page)
		if err != nil {
			return err
		}
		if len(sections) == 0 {
			fmt.Fprintf(os.Stderr, "No sections found for page %v\n", page)
			continue
		}

		for _, section := range desiredSections {
			contents, ok := sections[section]
			if !ok || len(contents.Packages) == 0 {
				continue
			}

			packageName := path.Base(contents.Packages[0])

			data, err := extractFileFromPackage(client, contents.File, packageName)
			if err != nil {
				return err
			}

			if err := writePage(section, page, data); err != nil {
				return err
			}
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dataHomeDir, err := dataHomeDir()
	if err != nil {
		return err
	}

	shortDataHomeDir := strings.Replace(dataHomeDir, homeDir, "~", 1)
	fmt.Fprintf(os.Stderr, finishMessage, shortDataHomeDir, dataHomeDir, dataHomeDir)

	return nil
}

func writePage(section, page string, data []byte) error {
	sectionPath, err := sectionPath(section)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sectionPath, 0o755); err != nil {
		return err
	}

	pagePath, err := pagePath(page, section)
	if err != nil {
		return err
	}
	if err := os.WriteFile(pagePath, data, 0o644); err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shortPagePath := strings.Replace(pagePath, homeDir, "~", 1)
	fmt.Fprintf(os.Stderr, "Downloaded %v(%v) to %v\n", strings.ToUpper(page), section, shortPagePath)

	return nil
}

func findSectionsForPage(client *deb.Client, page string) (map[string]deb.Contents, error) {
	var pagePaths []string
	sectionsByFile := make(map[string]string)

	for _, section := range DefaultSections() {
		path := fmt.Sprintf("usr/share/man/man%c/%v.%v.gz", section[0], page, section)
		pagePaths = append(pagePaths, path)
		sectionsByFile[path] = section
	}

	contents, err := client.QueryContents(pagePaths)
	if err != nil {
		return nil, err
	}

	results := make(map[string]deb.Contents)

	for _, content := range contents {
		section := sectionsByFile[content.File]
		results[section] = content
	}

	return results, nil
}

func extractFileFromPackage(client *deb.Client, filePath, packageName string) ([]byte, error) {
	pack, err := client.QueryPackage(packageName)
	if err != nil {
		return nil, err
	}

	download, err := client.Download(pack)
	if err != nil {
		return nil, err
	}
	defer download.Close()

	debDataFile, err := extractDebDataFile(download, filePath)
	if err != nil {
		return nil, fmt.Errorf("error extracting file: %w", err)
	}

	pageContents, err := io.ReadAll(debDataFile)
	if err != nil {
		return nil, err
	}

	return pageContents, nil
}

func extractDebDataFile(r io.Reader, file string) (io.Reader, error) {
	debData, err := extractDebData(r)
	if err != nil {
		return nil, fmt.Errorf("error extracting deb data: %w", err)
	}

	xzReader, err := xz.NewReader(debData)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(xzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "./"+file {
			return tarReader, nil
		}
	}

	return nil, fmt.Errorf("file not found in package: %v", file)
}

func extractDebData(r io.Reader) (io.Reader, error) {
	reader, err := ar.NewReader(r)
	if err != nil {
		return nil, err
	}

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "data.tar.xz" {
			return reader, nil
		}
	}

	return nil, fmt.Errorf("data.tar.xz not found")
}
