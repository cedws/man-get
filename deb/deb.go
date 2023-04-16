package deb

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

type option func(*Client)

func WithMirror(mirror string) option {
	return func(a *Client) {
		a.mirror = mirror
	}
}

func WithDistribution(dist string) option {
	return func(a *Client) {
		a.distribution = dist
	}
}

func WithArch(arch string) option {
	return func(a *Client) {
		a.arch = arch
	}
}

type Client struct {
	httpClient   *http.Client
	mirror       string
	distribution string
	arch         string
}

func NewAptClient(opts ...option) *Client {
	aptClient := Client{
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(&aptClient)
	}
	return &aptClient
}

// Download returns a ReadCloser representing a stream of the package's data from the mirror.
func (a *Client) Download(pack Package) (io.ReadCloser, error) {
	resp, err := a.httpClient.Get(fmt.Sprintf("%v/%v", a.mirror, pack.Filename))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// GetPakageByName returns a package by name.
func (a *Client) GetPackageByName(name string) (Package, error) {
	packageReader, err := a.Packages()
	if err != nil {
		return Package{}, err
	}
	for {
		pack, err := packageReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Package{}, err
		}
		if pack.Name == name {
			return pack, nil
		}
	}

	return Package{}, fmt.Errorf("package %v not found", name)
}

// Packages returns a PackageReader for the current distribution and architecture.
func (a *Client) Packages() (*PackageReader, error) {
	resp, err := a.httpClient.Get(fmt.Sprintf("%v/dists/%v/main/binary-%v/Packages.gz", a.mirror, a.distribution, a.arch))
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewPackageReader(reader), nil
}

// GetPackagesByFile returns the list of packages a given file path is contained in.
func (a *Client) GetPackagesByFile(file string) ([]string, error) {
	contentsReader, err := a.Contents()
	if err != nil {
		return nil, err
	}
	for {
		contents, err := contentsReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if contents.File == file {
			return contents.Packages, nil
		}
	}

	return nil, fmt.Errorf("file %v not found in any packages", file)
}

// Contents returns a ContentsReader for the current distribution and architecture.
func (a *Client) Contents() (*ContentsReader, error) {
	resp, err := a.httpClient.Get(fmt.Sprintf("%v/dists/%v/main/Contents-%v.gz", a.mirror, a.distribution, a.arch))
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewContentsReader(reader), nil
}
