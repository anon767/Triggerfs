# Configurable File System

An overlay for your FS using Go and Fuse native binding.
Call and hook into functions by accessing,opening,reading,... certain files.

# Config

create a config.json file, the structure of the contents should look like this:

```
{
// define a file
"/testfile": [{"permission":655, "exec":"echo foobar", "size":555, "ctime":1551220000, "mtime":1551220000 }],

// define a directory
// first entry defines directory
// all other entries will be treated as file definitions
// multiple entries are valid for directories only
"/testdir/": [
	{"permission":"0755", "ctime":1551220000, "mtime":1551220000},
	{"permission":755, "pattern":"*.txt", "exec":"echo foobar.txt", "size":500, "ctime":1551220000, "mtime":1551220000  },
	{"permission":755, "pattern":"*.pdf", "exec":"echo foobar.pdf", "size":500, "ctime":1551220000, "mtime":1551220000  }
	]
}
```

# Usage

```
go get
go build
mkdir mountpoint
./triggerfs mountpoint/ 
ls mountpoint
cat mountpoint/testfile

```

# Clean up
```
fusermount -u mnt
```
