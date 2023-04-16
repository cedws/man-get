package man

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/cedws/man1c/deb"
	"github.com/laher/argo/ar"
	"github.com/ulikunitz/xz"
)

const mandocRenderCmd = "mandoc"

func GetPages(section string, pages []string) {
	client := deb.NewAptClient(
		deb.WithMirror("https://ftp.debian.org/debian"),
		deb.WithDistribution("bullseye"),
		deb.WithArch("amd64"),
	)

	for _, page := range pages {
		manpage, err := getManpage(client, section, page)
		if err != nil {
			log.Fatal(err)
		}

		mandocCmd := exec.Command(mandocRenderCmd)
		pagerCmd := exec.Command(pager())

		mandocCmd.Stdin = bytes.NewBuffer(manpage)
		pagerCmd.Stdin, err = mandocCmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		pagerCmd.Stdout = os.Stdout
		pagerCmd.Stderr = os.Stderr

		if err := mandocCmd.Start(); err != nil {
			log.Fatal(err)
		}
		if err := pagerCmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

func pager() string {
	pager, ok := os.LookupEnv("MANPAGER")
	if ok {
		return pager
	}
	pager, ok = os.LookupEnv("PAGER")
	if ok {
		return pager
	}
	return "less"
}

func getManpage(client *deb.Client, section, page string) ([]byte, error) {
	manpagePath := fmt.Sprintf("usr/share/man/man%v/%v.%v.gz", section, page, section)
	pack, err := getFirstPackageByFile(client, manpagePath)
	if err != nil {
		return nil, err
	}

	download, err := client.Download(pack)
	if err != nil {
		return nil, err
	}
	defer download.Close()

	data, err := extractDebData(download)
	if err != nil {
		return nil, fmt.Errorf("error extracting deb package: %w", err)
	}

	manpage, err := extractDebDataFile(data, manpagePath)
	if err != nil {
		return nil, fmt.Errorf("error extracting file: %w", err)
	}

	return io.ReadAll(manpage)
}

func extractDebDataFile(r io.Reader, filepath string) (io.Reader, error) {
	xzReader, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(xzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("package doesn't contain file %v", filepath)
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "./"+filepath {
			gz, err := gzip.NewReader(tarReader)
			if err != nil {
				return nil, err
			}

			return gz, nil
		}
	}
}

func extractDebData(r io.Reader) (io.Reader, error) {
	reader, err := ar.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("error unpacking package: %w", err)
	}

	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("data.tar.xz not found")
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "data.tar.xz" {
			return reader, nil
		}
	}
}

func getFirstPackageByFile(client *deb.Client, file string) (deb.Package, error) {
	packages, err := client.GetPackagesByFile(file)
	if err != nil {
		return deb.Package{}, err
	}

	if len(packages) >= 1 {
		name := path.Base(packages[0])

		pack, err := client.GetPackageByName(name)
		if err != nil {
			return deb.Package{}, err
		}

		return pack, nil
	}

	return deb.Package{}, nil
}
