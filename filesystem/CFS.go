package filesystem

import (
	"path/filepath"
	"fmt"
	"strings"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

)

/*
Trigger File System

The Idea?

execute configurable commands on read to generate filecontent on the fly
*/

// shamelessly ripped from https://github.com/hanwen/go-fuse/blob/master/benchmark/statfs.go

type Conf struct {
	Pattern    string
	Exec       string
}

type triggerFS struct {
	pathfs.FileSystem
	entries map[string]*fuse.Attr
	dirs map[string][]fuse.DirEntry
	conf map[string][]Conf
}


func NewTriggerFS() *triggerFS {
	return &triggerFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		entries:    make(map[string]*fuse.Attr),
		dirs:       make(map[string][]fuse.DirEntry),
		conf:       make(map[string][]Conf),
	}
}


func (fs *triggerFS) Add(name string, permission uint32,pattern string, exec string, a *fuse.Attr) {
	name = strings.TrimRight(name, "/")
	_, ok := fs.entries[name]
	if ok {
		return
	}

	fs.entries[name] = a
	fs.conf[name] = append(fs.conf[name], Conf{Pattern: pattern, Exec: exec})
	if name == "/" || name == "" {
		return
	}

	dir, base := filepath.Split(name)
	dir = strings.TrimRight(dir, "/")
	fs.dirs[dir] = append(fs.dirs[dir], fuse.DirEntry{Name: base, Mode: a.Mode})
	//fs.conf[dir] = append(fs.conf[dir], Conf{Pattern: pattern, Exec: exec})
	fmt.Printf("v fs.dirs: %v\n", fs.dirs[dir])
	fs.Add(dir, permission, pattern, exec, &fuse.Attr{Mode: fuse.S_IFDIR | permission})
}


func (fs *triggerFS) AddFile(name string, permission uint32, pattern string, exec string) {
	fs.Add(name, permission, pattern, exec, &fuse.Attr{Mode: fuse.S_IFREG | permission})
}


func (fs *triggerFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if d := fs.dirs[name]; d != nil {
		fmt.Printf("dirs %s: %v\n", name, context)
		//return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
		return fs.entries[name], fuse.OK
	}
	if name == "" {
		fmt.Printf("dirs %s: %v\n", name, context)
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	}
	//not found
	fmt.Printf("dirs %s: %v\n", name, context)
	return nil, fuse.ENOENT
}

func (fs *triggerFS) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	entries := fs.dirs[name]
	if entries == nil {
		return nil, fuse.ENOENT
	}
	return entries, fuse.OK
}


func (fs *triggerFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	fmt.Println("Open called")
	return fs.FileSystem.Open(name, flags, context)
}




