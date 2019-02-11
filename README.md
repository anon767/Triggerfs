# Configurable File System

An overlay for your FS using Go and Fuse native binding.
Call and hook into functions by accessing,opening,reading,... certain files.

# Config

create a config.json file, the structure of the contents should look like this:

```
[
  {
    "permission": "0777",
    "Pattern": ".*",
    "Path": "test.txt",
    "Exec": "/home/tom/go/src/configurablefs/bla.sh"
  }
]
```

# Usage

```
go get
go build
mkdir mountpoint
mkdir test
./configurablefs test/ mountpoint/ &
cd mountpoint
echo test > test
cat test
```

# Clean up
```
cd ..
sudo umount mountpoint
```