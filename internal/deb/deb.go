package deb

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"slices"
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

func WithCacheDir(cacheDir string) option {
	return func(a *Client) {
		a.cacheDir = cacheDir
	}
}

type Client struct {
	httpClient   *http.Client
	mirror       string
	distribution string
	arch         string
	cacheDir     string
}

func NewAptClient(opts ...option) *Client {
	aptClient := Client{
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(&aptClient)
	}

	if aptClient.cacheDir != "" {
		aptClient.httpClient.Transport = &cachingRoundTripper{
			RoundTripper: http.DefaultTransport,
			cacheDir:     aptClient.cacheDir,
		}
	}

	return &aptClient
}

// Download returns a ReadCloser representing a stream of the package's data from the mirror.
func (a *Client) Download(pack Package) (io.ReadCloser, error) {
	url := fmt.Sprintf("%v/%v", a.mirror, pack.Filename)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// QueryPackage returns a package queried by name.
func (a *Client) QueryPackage(name string) (Package, error) {
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
	url := fmt.Sprintf("%v/dists/%v/main/binary-%v/Packages.gz", a.mirror, a.distribution, a.arch)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewPackageReader(reader), nil
}

// QueryContents returns a list of packages that contain the given files.
func (a *Client) QueryContents(files []string) ([]Contents, error) {
	var contents []Contents

	contentsReader, err := a.Contents()
	if err != nil {
		return nil, err
	}
	for {
		entry, err := contentsReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if slices.Contains(files, entry.File) {
			contents = append(contents, entry)
		}
	}

	return contents, nil
}

// Contents returns a ContentsReader for the current distribution and architecture.
func (a *Client) Contents() (*ContentsReader, error) {
	url := fmt.Sprintf("%v/dists/%v/main/Contents-%v.gz", a.mirror, a.distribution, a.arch)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewContentsReader(reader), nil
}
