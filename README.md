# tile-dump

This is a hacked-up tool using [tegola](github.com/go-spatial/tegola) to dump
the contents of a slippy tile into ~wkt~ a Go literal.

The vendor directory contains a modifed version of the tegola repository
with the geom package removed from its vendor directory. This prevents type
mismatches since the geom package is used in this repository.

## usage

```bash
$ tile-dump [tegola-config.toml] [z/x/y]
```

The output goes to a file named `{map_name}_{z}_{x}_{y}.go`
