package deb

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

const (
	etag          = "etag"
	contentLength = "content-length"
)

type cachingRoundTripper struct {
	http.RoundTripper
	cacheDir string
}

func (c *cachingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	etag, err := strconv.Unquote(resp.Header.Get(etag))
	if err != nil {
		return nil, err
	}
	length, err := strconv.ParseInt(resp.Header.Get(contentLength), 10, 64)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return nil, err
	}

	path := path.Join(c.cacheDir, etag)
	stat, statErr := os.Stat(path)

	opts := os.O_CREATE | os.O_RDWR
	if statErr == nil && stat.Size() != length {
		// file exists but is wrong size, maybe it was only partially downloaded
		// truncate it and redownload later
		opts |= os.O_TRUNC
	}

	file, err := os.OpenFile(path, opts, 0o644)
	if err != nil {
		return nil, err
	}
	if os.IsNotExist(statErr) || stat.Size() != length {
		// file doesn't exist or was partially downloaded
		if _, err := io.Copy(file, resp.Body); err != nil {
			return nil, err
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	resp.Body = file
	return resp, nil
}
