package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"fmt"
	"log"
	
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	
	"github.com/davecgh/go-spew/spew"
)

func main() {

	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  triggerfs MOUNTPOINT")
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
				attr, permission := parser.ConfigToAttr(event[i], true)
				fmt.Printf("adddir: %+v\n", attr)
				fs.Add(path, uint32(permission), event[i].Pattern, event[i].Exec, attr)
			}
		} else {
			// file
			// only one entry per file definition allowed
			attr, permission := parser.ConfigToAttr(event[0], false)
			fmt.Printf("addfile: %+v\n", attr)
			fs.AddFile(path, uint32(permission), event[0].Pattern, event[0].Exec, attr)
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
