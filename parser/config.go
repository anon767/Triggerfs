package parser

import (
	"log"
	"strconv" 
	"github.com/lytics/confl"
	"github.com/hanwen/go-fuse/fuse"
)

type Entry struct {
	Permission string  `confl:"permission"`
	Exec       string `confl:"exec"`
	Mtime      int  `confl:"mtime"`
	Ctime      int  `confl:"ctime"`
	Atime      int  `confl:"atime"`
	Size       int  `confl:"size"`
}


type Config struct {
	// triggerFS config
	Title string `confl:"title"`
	Caching bool `confl:"size_cache"`
	PrebuildCache bool `confl:"prebuild_cache"`
	//entries
	File map[string]Entry
	Dir map[string]Entry
	Pattern map[string]Entry
	
}


//type Config map[string][]Entry

func Parseconfig(configFile string) (config Config) {
	
	var cfg Config
	_, err := confl.DecodeFile(configFile, &cfg)
	if err != nil {
	  log.Println("error decoding config: ",err)
	}

	return cfg
}


func ConfigToAttr(config Entry, dir bool) (*fuse.Attr) {
	attr := &fuse.Attr{}
	permission := uint32(0644)
	mode := uint32(fuse.S_IFREG)
	
	if dir {
		permission = 0755
		mode = fuse.S_IFDIR
	}
	
	if config.Permission != "" {
		int_permission, err := strconv.ParseUint(config.Permission, 8, 32)
		if err == nil {
			permission = uint32(int_permission)
		}
		
	}
	attr.Mode = mode | permission

	attr.Size  = uint64(config.Size)
	attr.Mtime = uint64(config.Mtime)
	attr.Atime = uint64(config.Atime)
	attr.Ctime = uint64(config.Ctime)

	
	return attr
}
