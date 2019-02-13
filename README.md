# Configurable File System

An overlay for your FS using Go and Fuse native binding.
Call and hook into functions by accessing,opening,reading,... certain files.

# Config

create a config.json file, the structure of the contents should look like this:

```
{
// define a file
"/testfile": [{"path":"/testfile", "permission":655, "exec":"echo foobar" }],

// define a directory
// multiple entries are valid for directories only
"/testdir/": [
	{"path":"/testdir", "permission":755, "pattern":"*.txt", "exec":"echo foobar.txt" },
	{"path":"/testdir", "permission":755, "pattern":"*.pdf", "exec":"echo foobar.pdf" }
	]
}
```

# Usage

```
go get
go build
mkdir mountpoint
./configurablefs mountpoint/ 
ls mountpoint
cat mountpoint/testfile

```

# Clean up
```
fusermount -u mnt
```
