# TriggerFS

An overlay for your FS using Go and Fuse native binding.
Execute configurable commands on read calls of files or patterns of filenames.

# Config

create a config.json file, the structure of the contents should look like this:
```
{
// define a file
// i.e. "foobar" as content of testfile
"/testfile": [{"permission":655, "exec":"echo foobar", "size":555, "ctime":1551220000, "mtime":1551220000 }],

// define a directory
// multiple entries are valid for directories only
// first entry defines directory
// all other entries will be treated as file definitions
// i.e. "foobar" as content of all *.txt files that are accessed
"/testdir/": [
	{"permission":"0755", "ctime":1551220000, "mtime":1551220000},
	{"permission":755, "pattern":".txt", "exec":"echo foobar", "size":500, "ctime":1551220000, "mtime":1551220000  }
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
fusermount -u mountpoint
```
