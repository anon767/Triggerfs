package main

import (
	"configurablefs/filesystem"
	"configurablefs/parser"
	"flag"
	"fmt"
	"log"
	"strconv" 
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
		fmt.Printf("event: %v\n", event)
		fmt.Println(path + "\n")
		if path[len(path)-1] == '/' {
			// directory
			// multiple entries possible
			for i := 0; i < len(event); i++ {
				//there must be a  better way
				var permission uint32
				
				int_permission, err := strconv.Atoi(event[i].Permission) 
				if err != nil {
					permission = uint32(0755)
				} else {
					permission = uint32(int_permission)
				}
				fmt.Println("adddir: " + string(permission) + "\n")
				fs.Add(path, permission, &fuse.Attr{Mode: fuse.S_IFDIR | permission})
			}
		} else {
			// file
			// only one entry per file definition allowed
			int_permission, err := strconv.Atoi(event[0].Permission)
			var permission uint32
			if err != nil {
				permission = uint32(0655)
			} else {
				permission = uint32(int_permission)
			}
			fmt.Println("addfile: " + string(permission) + "\n")
			fs.AddFile(path, permission)
		}
	}
	
	spew.Dump(fs)
	//fmt.Printf("%v\n",fs)
	
	mountpoint := flag.Arg(0)
	
	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	fmt.Printf("%s is mountpoint\n", mountpoint)

	server.Serve()
}
