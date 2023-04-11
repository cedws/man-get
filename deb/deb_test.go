package deb

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const packages = `
Package: zip
Priority: optional
Section: utils
Installed-Size: 100

Package: unzip
Priority: optional
Section: utils
Installed-Size: 100
`

func TestPackageReader(t *testing.T) {
	r := strings.NewReader(packages)
	reader := NewPackageReader(r)

	var packages []Package
	for {
		pack, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			assert.Fail(t, err.Error())
		}

		packages = append(packages, pack)
	}

	assert.Equal(t, "zip", packages[0].Name)
	assert.Equal(t, "unzip", packages[1].Name)
}
