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
	nextinode int
}


func NewTriggerFS() *triggerFS {
	return &triggerFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		entries:    make(map[string]*fuse.Attr),
		dirs:       make(map[string][]fuse.DirEntry),
		conf:       make(map[string][]Conf),
		nextinode:	0,
	}
}


func PrepareCmd(command string, path string, file string) string {
	exec := strings.Replace(command, "%FILE%", file, -1)
	exec = strings.Replace(exec, "%PATH%", path, -1)
	return exec
}


func ExecCmd(command string) string {
	fmt.Printf("Executing: %s\n", command)
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


func UpdateSize(attr *fuse.Attr, size int) *fuse.Attr {
	attr.Size = uint64(size)
	return attr
}


func (fs *triggerFS) GetNextInode() int {
	fs.nextinode++
	return fs.nextinode
}



func (fs *triggerFS) Add(name string, pattern string, exec string, attr *fuse.Attr) {
	//name = strings.TrimRight(name, "/")
	//name = "/" + name
	if name == "" {
		name = "/"
	}
	fs.conf[name] = append(fs.conf[name], Conf{Pattern: pattern, Exec: exec, Attr: attr})
	
	if fs.entries[name] != nil {
		return
	}
	//attr.Inode = fs.GetNextInode()

	fs.entries[name] = attr
	fmt.Printf("ADD: %s\n", name)
	
	
	dir, base := filepath.Split(name)
	if dir != "/" {
		dir = strings.TrimRight(dir, "/")
	}
	if base != "" {
		fs.dirs[dir] = append(fs.dirs[dir], fuse.DirEntry{Name: base, Mode: attr.Mode})
	}
	if name == "/" {
		return
	}
	
	fs.Add(dir, "", "", &fuse.Attr{Mode: fuse.S_IFDIR | 0755, Size: 4096})
}


func (fs *triggerFS) Deletable() bool {
	return false
}


func (fs *triggerFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	name = "/" + name
	//if name == "/" {
		////fmt.Printf("getattr name empty %s: %v\n", name, context)
		//return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	//}
	if d := fs.entries[name]; d != nil {
		//fmt.Printf("getattr found %s: %v\n", name, context)
		return fs.entries[name], fuse.OK
	}
	
	dirname, filename := filepath.Split(string(name))
	if dirname != "/" {
		dirname = strings.TrimRight(dirname, "/")
	}
	cfg := fs.conf[dirname]
	if cfg != nil {
		for i := 1; i < len(cfg); i++ {
			if MatchFile(filename, cfg[i].Pattern) {
				//fmt.Printf("getattr found dir rule %s: %v\n", name, cfg[i].Attr)
				return cfg[i].Attr, fuse.OK
				
			}
		}
	}
	//not found
	//fmt.Printf("getattr not found %s: %v\n", name, context)
	return nil, fuse.ENOENT
}

func (fs *triggerFS) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	//name = name + "/" 
	name = "/" + name
	
	entries := fs.dirs[name]
	//fmt.Printf("Opendir: %s: %v\n", name, entries)
	//if entries == nil {
		//return nil, fuse.ENOENT
	//}
	return entries, fuse.OK
}


func (fs *triggerFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	name = "/" + name
	fmt.Printf("open not found: %s\n", name)
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	
	dirname, filename := filepath.Split(string(name))
	if dirname != "/" {
		dirname = strings.TrimRight(dirname, "/")
	}
	//match file
	cfg := fs.conf[name]
	if cfg != nil {
		exec := PrepareCmd(cfg[0].Exec, name, filename)
		fmt.Printf("Open file: %s\n",name)
		
		content := ExecCmd(exec)
		//fs.entries[name] = UpdateSize(fs.entries[name], len(content))
		return nodefs.NewDataFile([]byte(content)), fuse.OK
	}
	
	//match dir
	cfg = fs.conf[dirname]
	if cfg != nil {
		fmt.Println("matched dir")
		for i := 1; i < len(cfg); i++ {
			if MatchFile(filename, cfg[i].Pattern) {
				exec := PrepareCmd(cfg[i].Exec, name, filename)
				fmt.Printf("Open match dir: %s\n",name)
				
				content := ExecCmd(exec)
				
				// resetting the size of matched files. maybe we should do a fs.Add() here to index the called file
				//fmt.Printf("Old Size of %s: %i\n",name,int(cfg[i].Attr.Size))
				//attr := cfg[i].Attr
				//attr.Size = uint64(len(content))
				////fs.conf[dirname][i].Attr = attr
				//fs.conf[dirname][i].Attr = UpdateSize(fs.conf[dirname][i].Attr, len(content))
				//fs.entries[name] = UpdateSize(fs.conf[dirname][i].Attr, len(content))
				//fmt.Printf("New Size of %s: %i\n",name,int(cfg[i].Attr.Size))
				
				return nodefs.NewDataFile([]byte(content)), fuse.OK
			}
		}
	}
	//not found
	fmt.Printf("open not found: %s\n", name)
	return nil, fuse.ENOENT
}


