package filesystem

import (
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

/*
Configurable File System (CFS)

The Idea?

act as Loopback FS and eventually call a hooked method depending on maaybe config file?
*/

type CFS struct {
	pathfs.FileSystem
}

func (me *CFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {

	fmt.Print("Here can I call Stuff")

	return me.FileSystem.OpenDir(name, context)
}

func (me *CFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	fmt.Print("Here can I call Stuff")
	return me.FileSystem.Open(name, flags, context)
}
