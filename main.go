package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"fmt"
	"log"
	
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

)

func main() {
	
	//configfile := "config.json"
	var configfile string
	//fmt.Printf("%v\n", flag.Args())
	flag.StringVar(&configfile,"c", "config.json", "Configfile")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  triggerfs MOUNTPOINT")
	}
	mountpoint := flag.Arg(0)
	
	
	
	fmt.Printf("reading config:%s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	
	fs := filesystem.NewTriggerFS()
	for path, event := range config {
		fmt.Printf("event: %v\n", event)
		fmt.Println(path + "\n")
		if path[len(path)-1] == '/' {
			// directory
			attr, permission := parser.ConfigToAttr(event[0], true)
			fs.Add(path, uint32(permission), event[0].Pattern, event[0].Exec, attr)
			// multiple entries possible
			for i := 1; i < len(event); i++ {
				attr, permission := parser.ConfigToAttr(event[i], false)
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
	
	//fmt.Printf("%v\n",fs)
	
	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	fmt.Printf("%s is mountpoint\n", mountpoint)

	server.Serve()
}
