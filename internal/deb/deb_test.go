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

const contents = `
bin/afio                                                utils/afio
bin/bash                                                shells/bash
bin/bash-static                                         shells/bash-static
bin/brltty                                              admin/brltty
bin/bsd-csh                                             shells/csh
bin/btrfs                                               admin/btrfs-progs
bin/btrfs-convert                                       admin/btrfs-progs
bin/btrfs-find-root                                     admin/btrfs-progs
bin/btrfs-image                                         admin/btrfs-progs
bin/btrfs-map-logical                                   admin/btrfs-progs
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

func TestContentsReader(t *testing.T) {
	r := strings.NewReader(contents)
	reader := NewContentsReader(r)

	var contents []Contents
	for {
		entry, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			assert.Fail(t, err.Error())
		}

		contents = append(contents, entry)
	}

	assert.Equal(t, "bin/afio", contents[0].File)
	assert.Equal(t, "bin/bash", contents[1].File)
	assert.Equal(t, []string{"utils/afio"}, contents[0].Packages)
	assert.Equal(t, []string{"shells/bash"}, contents[1].Packages)
}
