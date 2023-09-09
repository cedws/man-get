# man-get
CLI tool to grab Debian manpages. I put this together out of frustration of either having to browse the web or install tools inside Docker to see manpages when working on a non-Linux system.

## Usage
man-get behaves similarly to man. When a section is unspecified, it will download all available sections for the given pages. You can open these sections with `man` after appending the data path to the `MANPATH` environment variable. Since it supports the XDG Base Directory spec, you can override this path by setting `XDG_DATA_HOME`.

man-get caches fetched indexes from the Debian package mirror to be friendly to their servers. The cache path can also be overridden by setting `XDG_CACHE_HOME`.

```
$ man-get tar
Downloaded TAR(1) to ~/.local/share/man-get/man1/tar.1.gz
Downloaded TAR(5) to ~/.local/share/man-get/man5/tar.5.gz
Append ~/.local/share/man-get to your MANPATH to open the downloaded manpages:

        Bash/Zsh:
        $ export MANPATH="$MANPATH:/Users/connor.edwards/.local/share/man-get"

        Fish:
        $ set -x MANPATH "$MANPATH:/Users/connor.edwards/.local/share/man-get"
$ man-get 1 ed
Downloaded ED(1) to ~/.local/share/man-get/man1/ed.1.gz
Append ~/.local/share/man-get to your MANPATH to open the downloaded manpages:

        Bash/Zsh:
        $ export MANPATH="$MANPATH:/Users/connor.edwards/.local/share/man-get"

        Fish:
        $ set -x MANPATH "$MANPATH:/Users/connor.edwards/.local/share/man-get"
$ export MANPATH="$MANPATH:/Users/connor.edwards/.local/share/man-get"
$ man tar
$ man ed
```