package filesystem

import (
	"path/filepath"
	"triggerfs/parser"
	"log"
	"os/exec"
	
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


type Cache struct {
	// content should be cached in sqlite in the future
	//Content		string
	Attr		*fuse.Attr
}

type triggerFS struct {
	pathfs.FileSystem
	entries map[string]*fuse.Attr
	dirs map[string][]fuse.DirEntry
	conf map[string][]Conf
	BaseConf map[string]parser.Config
	cache map[string]*fuse.Attr
	LogLevel int
//	nextinode int
}


func NewTriggerFS() *triggerFS {
	return &triggerFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		entries:    make(map[string]*fuse.Attr),
		dirs:       make(map[string][]fuse.DirEntry),
		conf:       make(map[string][]Conf),
		BaseConf:   make(map[string]parser.Config),
		cache:      make(map[string]*fuse.Attr),
		LogLevel:	0,
//		nextinode:	0,
	}
}


//func (fs *triggerFS) GetNextInode() int {
	//fs.nextinode++
	//return fs.nextinode
//}


func (fs *triggerFS) AddFile(name string, exec string, attr *fuse.Attr) {
	if fs.entries[name] != nil {
		return
	}
	
	if fs.LogLevel >= 1 {
		log.Println("Adding File: ", name)
	}
	
	dir, base := filepath.Split(name)
	dir = strings.TrimRight(dir, "/")
	// run exec command to prebuild cache if enabled
	if fs.BaseConf["triggerFS"].PrebuildCache {
		if fs.cache[name] != nil {
			cmd := PrepareCmd(exec, name, base)
			content := ExecCmd(cmd)
			fs.cache[name] = UpdateSize(attr, len(content))
		}
	}
	
	fs.conf[name] = append(fs.conf[name], Conf{Pattern: "", Exec: exec, Attr: attr})
	fs.entries[name] = attr
	fs.dirs[dir] = append(fs.dirs[dir], fuse.DirEntry{Name: base, Mode: attr.Mode})
	if fs.LogLevel >= 3 {
		log.Printf("Adding %s to %s: %v\n", base, dir, fs.dirs[dir])
	}
	
	dirattr := &fuse.Attr{
		Mode: fuse.S_IFDIR | 0755,
		Size: 4096,
		Mtime: attr.Mtime,
		Ctime: attr.Ctime,
		Atime: attr.Atime}
	fs.AddDir(dir, dirattr)
	
}


func (fs *triggerFS) AddDir(name string, attr *fuse.Attr) {
	name = strings.TrimRight(name, "/")
		
	if fs.entries[name] != nil {
		return
	}
	
	if fs.LogLevel >= 1 {
		log.Printf("Adding Dir: %s\n", name)
	}
	
	fs.entries[name] = attr
	dir, base := filepath.Split(name)
	dir = strings.TrimRight(dir, "/")

	if dir == name {
		return
	}
	fs.dirs[dir] = append(fs.dirs[dir], fuse.DirEntry{Name: base, Mode: attr.Mode})
	
	fs.AddDir(dir, attr)
	
}


func (fs *triggerFS) AddPattern(name string, exec string, attr *fuse.Attr) {
	if fs.LogLevel >= 1 {
		log.Println("Adding Pattern: ", name)
	}
	
	dir, base := filepath.Split(name)
	dir = strings.TrimRight(dir, "/")

	fs.conf[dir] = append(fs.conf[dir], Conf{Pattern: base, Exec: exec, Attr: attr})
	
	dirattr := &fuse.Attr{
		Mode: fuse.S_IFDIR | 0755,
		Size: 4096,
		Mtime: attr.Mtime,
		Ctime: attr.Ctime,
		Atime: attr.Atime}

	fs.AddDir(dir, dirattr)
}


func (fs *triggerFS) CacheFileAttr(name string, attr *fuse.Attr, size int) bool {
	if fs.BaseConf["triggerFS"].Caching {
		attr.Size = uint64(size)
		fs.cache[name] = attr
		if fs.LogLevel >= 3 {
			log.Printf("caching attributes of file %s", name)
		}
		return true
	}
	return false
		
}


func (fs *triggerFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	// return cached attributes if it exist
	if attr, ok := fs.cache[name]; ok {
		if fs.LogLevel >= 2 {
			log.Printf("getattr found cache %s: %v\n", name, attr)
		}
		return attr, fuse.OK
	}
	
	if d := fs.entries[name]; d != nil {
		if fs.LogLevel >= 2 {
			log.Printf("getattr found file %s: %v\n", name, fs.entries[name])
		}
		return fs.entries[name], fuse.OK
	}
	
	dirname, filename := filepath.Split(string(name))
	dirname = strings.TrimRight(dirname, "/")

	cfg := fs.conf[dirname]
	if cfg != nil {
		for i := 0; i < len(cfg); i++ {
			if cfg[i].Pattern == "" {
				continue
			}
			if MatchFile(filename, cfg[i].Pattern) {
				if fs.LogLevel >= 2 {
					log.Printf("getattr found pattern for %s: %v\n", name, cfg[i].Attr)
				}
				return cfg[i].Attr, fuse.OK
				
			}
		}
	}
	//not found
	if fs.LogLevel >= 3 {
		log.Printf("getattr not found %s: %v\n", name, context)
	}
	return nil, fuse.ENOENT
}


func (fs *triggerFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {

	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	dirname, filename := filepath.Split(string(name))
	dirname = strings.TrimRight(dirname, "/")

	//match file
	cfg := fs.conf[name]
	if cfg != nil {
		exec := PrepareCmd(cfg[0].Exec, name, filename)
		if fs.LogLevel >= 2 {
			log.Printf("open file %s with command: %s\n", name, exec)
		}
		
		content := ExecCmd(exec)
		// resetting the size attribute because some programs are strict about the size given by getattr() to be the actual content size
		// it depends on the given exec command for it to work as intended
		fs.CacheFileAttr(name, fs.entries[name], len(content))
		return nodefs.NewDataFile([]byte(content)), fuse.OK
	}
	
	//match dir
	cfg = fs.conf[dirname]
	if cfg != nil {
		for i := 0; i < len(cfg); i++ {
			if MatchFile(filename, cfg[i].Pattern) {
				exec := PrepareCmd(cfg[i].Exec, name, filename)
				if fs.LogLevel >= 2 {
					log.Printf("open file %s through pattern %s with command: %s\n", name, cfg[i].Pattern, exec)
				}
				content := ExecCmd(exec)
				
				// resetting the size of matched files. maybe we should do a fs.AddFile() here to index the called file
				// some programs are strict about the size given by getattr() to be the actual content size
				// so we cache it to have the correct size at the second open request at least (depending on the exec command of cause)
				fs.CacheFileAttr(name, fs.conf[dirname][i].Attr, len(content))
				
				// add matched file to fs tree after being opened once if enabled
				if fs.BaseConf["triggerFS"].UpdateTree {
									
					fs.AddFile(name, cfg[i].Exec, UpdateSize(fs.conf[dirname][i].Attr, len(content)))
					if fs.LogLevel >= 2 {
						log.Printf("added %s to fs tree in %s: %v\n", name, dirname, fs.dirs[dirname])
					}
				}
				//log.Printf("New Size of %s: %i\n",name,int(cfg[i].Attr.Size))
				return nodefs.NewDataFile([]byte(content)), fuse.OK
			}
		}
	}
	//not found
	if fs.LogLevel >= 3 {
		log.Printf("open file not found: %s\n", name)
	}
	return nil, fuse.ENOENT
}


func (fs *triggerFS) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {

	entries := fs.dirs[name]
	if fs.LogLevel >= 3 {
		log.Printf("open dir: %v\n", entries)
	}
	return entries, fuse.OK
}


func (fs *triggerFS) Deletable() bool {
	return false
}


func (fs *triggerFS) String() string {
	return fs.BaseConf["triggerFS"].Title
}


func PrepareCmd(command string, path string, file string) string {
	exec := strings.Replace(command, "%FILE%", file, -1)
	exec = strings.Replace(exec, "%PATH%", path, -1)
	return exec
}


func ExecCmd(command string) string {
	if fs.LogLevel >= 3 {
		log.Printf("executing: %s\n", command)
	}
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

