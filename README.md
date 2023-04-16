# man1c
Cross-platform CLI tool to grab Debian manpages. I quickly put this together out of frustration of either having to browse the web or install tools inside Docker images to see their manpges. At the moment it's terribly because there's no caching and it didn't try to make it at all efficient. Also, the name sucks, got a better one?

There's a tool called [debiman](https://github.com/Debian/debiman) from the Debian project themselves but it seems to be focused more on generating static documentation and is not cross-platform.

## Usage
This tool requires `mandoc` to be present. By default, it will use `$MANPAGER` as the pager, falling back to `$PAGER`, and finally to `less`.

```
$ man1c 5 tar
$ man1c 1 ed
```