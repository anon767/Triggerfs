package filesystem

import (
	"path/filepath"
	"fmt"
	"os/exec"
	"log"
	"strings"
	"regexp"
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
	Attr       *fuse.Attr
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


func ExecCmd(command string) string {
	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}


func MatchFile(file string, pattern string) bool { 
	matched, err := regexp.MatchString(pattern, file)
	if err != nil {
		log.Fatal(err)
	}
	if matched {
		return true
	}
	return false
}


func (fs *triggerFS) Add(name string, permission uint32,pattern string, exec string, attr *fuse.Attr) {
	//name = strings.TrimRight(name, "/")
	fs.conf[name] = append(fs.conf[name], Conf{Pattern: pattern, Exec: exec, Attr: attr})
	_, ok := fs.entries[name]
	if ok {
		return
	}

	fs.entries[name] = attr
	
	if name == "/" || name == "" {
		return
	}

	dir, base := filepath.Split(name)
	dir = strings.TrimRight(dir, "/")
	fs.dirs[dir] = append(fs.dirs[dir], fuse.DirEntry{Name: base, Mode: attr.Mode})
	fs.Add(dir, 0, "", "", &fuse.Attr{Mode: fuse.S_IFDIR | permission})
}


func (fs *triggerFS) AddFile(name string, permission uint32, pattern string, exec string, attr *fuse.Attr) {
	fs.Add(name, permission, pattern, exec, attr)
}


func (fs *triggerFS) Deletable() bool {
	return false
}

func (fs *triggerFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	name = "/" + name
	if name == "/" {
		fmt.Printf("getattr name empty %s: %v\n", name, context)
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	}
	if d := fs.entries[name]; d != nil {
		fmt.Printf("getattr found %s: %v\n", name, context)
		//return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
		return fs.entries[name], fuse.OK
	}
	
	dirname, filename := filepath.Split(string(name))
	cfg := fs.conf[dirname]
	if cfg != nil {
		for i := 1; i < len(cfg); i++ {
			if MatchFile(filename, cfg[i].Pattern) {
				fmt.Printf("getattr found dir rule %s: %v\n", name, cfg[i].Attr)
				return cfg[i].Attr, fuse.OK
				
			}
		}
	}
	//not found
	fmt.Printf("getattr not found %s: %v\n", name, context)
	return nil, fuse.ENOENT
}

func (fs *triggerFS) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	name = "/" + name
	entries := fs.dirs[name]
	if entries == nil {
		return nil, fuse.ENOENT
	}
	return entries, fuse.OK
}


func (fs *triggerFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	name = "/" + name
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	
	dirname, filename := filepath.Split(string(name))
	
	//match file
	cfg := fs.conf[name]
	if cfg != nil {
		exec := strings.Replace(cfg[0].Exec, "%FILE%", filename, -1)
		exec = strings.Replace(exec, "%PATH%", name, -1)
		fmt.Printf("Open file: %s -- %s\n",name, exec)
		content := ExecCmd(exec)
		
		//fmt.Printf("Open file: %s\n",name)
		//content := ExecCmd(cfg[0].Exec)
		return nodefs.NewDataFile([]byte(content)), fuse.OK
	}
	
	//match dir
	cfg = fs.conf[dirname]
	if cfg != nil {
		for i := 1; i < len(cfg); i++ {
			if MatchFile(filename, cfg[i].Pattern) {
				exec := strings.Replace(cfg[i].Exec, "%FILE%", filename, -1)
				exec = strings.Replace(exec, "%PATH%", name, -1)
				fmt.Printf("Open match dir: %s -- %s\n",name, exec)
				content := ExecCmd(exec)
				return nodefs.NewDataFile([]byte(content)), fuse.OK
			}
		}
	}
	//not found
	return nil, fuse.ENOENT
}


