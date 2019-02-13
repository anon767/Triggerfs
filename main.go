package main

import (
	"configurablefs/filesystem"
	"configurablefs/parser"
	"flag"
	"fmt"
	"log"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/davecgh/go-spew/spew"
)

func main() {

	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  configurablefs MOUNTPOINT")
	}
	
	fmt.Println("reading config:\n")
	config := parser.Parseconfig("config.json")
	spew.Dump(config)
	
	fs := filesystem.NewTriggerFS()
	
	for path, event := range config {
		
		if path[len(path)-1] == '/' {
			// directory
			// multiple entries possible
			for i := 0; i < len(event); i++ {
				permission := uint32(0755)
				if len(string(event[i].Permission)) < 3 { //there must be a  better way
					permission = uint32(event[i].Permission)
				}
				fs.Add(path, permission, &fuse.Attr{Mode: fuse.S_IFDIR | permission})
			}
		} else {
			// file
			// only one entry per file definition allowed
			fs.AddFile(path, config[path][0].Permission)
		}
	}
	
	spew.Dump(fs)
	
	mountpoint := flag.Arg(0)
	
	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	fmt.Printf("%s is mountpoint\n", mountpoint)

	server.Serve()
}
