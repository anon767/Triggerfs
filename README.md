# Configurable File System

An overlay for your FS using Go and Fuse native binding.
Call and hook into functions by accessing,opening,reading,... certain files.


# Usage

```
go get
go build
mkdir mountpoint
mkdir test
./configurablefs test/ mountpoint/ &
cd test
ls -la
cd ..
sudo umount mountpoint
```
