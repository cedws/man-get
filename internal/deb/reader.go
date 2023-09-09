package deb

import (
	"bufio"
	"io"
	"strings"
)

type PackageReader struct {
	scanner *bufio.Scanner
}

type ContentsReader struct {
	scanner *bufio.Scanner
}

type Package struct {
	Name     string
	Source   string
	Filename string
}

type Contents struct {
	File     string
	Packages []string
}

func NewPackageReader(r io.Reader) *PackageReader {
	const bufferSize = 128 * 1024

	// TODO: unfuckify this
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, bufferSize), bufferSize)

	return &PackageReader{scan}
}

func NewContentsReader(r io.Reader) *ContentsReader {
	return &ContentsReader{bufio.NewScanner(r)}
}

func (p *PackageReader) Next() (Package, error) {
	var pack Package

	// keep scanning until we get a non-empty line
	for {
		ok := p.scanner.Scan()
		if !ok {
			if err := p.scanner.Err(); err != nil {
				return Package{}, err
			}
			return Package{}, io.EOF
		}

		if p.scanner.Text() != "" {
			break
		}
	}

	for {
		line := p.scanner.Text()
		if line == "" {
			// end of this package entry
			return pack, nil
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := fields[0]
			value := fields[1]

			switch key {
			case "Package:":
				pack.Name = value
			case "Filename:":
				pack.Filename = value
			}
		}

		ok := p.scanner.Scan()
		if !ok {
			if err := p.scanner.Err(); err != nil {
				return Package{}, err
			}
			return pack, nil
		}
	}
}

func (c *ContentsReader) Next() (Contents, error) {
	for {
		ok := c.scanner.Scan()
		if !ok {
			if err := c.scanner.Err(); err != nil {
				return Contents{}, err
			}
			return Contents{}, io.EOF
		}

		line := c.scanner.Text()
		fields := strings.Fields(line)

		if len(fields) == 2 {
			return Contents{
				File:     fields[0],
				Packages: strings.Split(fields[1], ","),
			}, nil
		}
	}
}
