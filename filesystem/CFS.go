package filesystem

import (
	"configurablefs/parser"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

/*
Configurable File System (CFS)

The Idea?

act as Loopback FS and eventually call a hooked method depending on maybe config file?
*/

type CFS struct {
	pathfs.FileSystem
}

func (me *CFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	events := parser.Parseconfig("test.json")
	exec, matched := parser.MatchFile(name, events)
	if matched {
		return nodefs.NewDataFile([]byte(parser.ExecFile(exec))), fuse.OK
	}
	//base case
	return me.FileSystem.Open(name, flags, context)
}
