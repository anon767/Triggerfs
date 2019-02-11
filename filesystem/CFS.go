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
	events := parser.Parseconfig("config.json")
	matchedEvent, matched := parser.EventsMatchFile(name, events)
	if matched {
		return nodefs.NewDataFile([]byte(matchedEvent.ExecCmd(name))), fuse.OK
	}
	//base case
	return me.FileSystem.Open(name, flags, context)
}
