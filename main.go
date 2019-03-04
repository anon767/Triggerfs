package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"fmt"
	"strings"
	"log"
	
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

)

func main() {
	
	//configfile := "config.json"
	var configfile string
	flag.StringVar(&configfile,"c", "config.json", "Configfile")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  triggerfs MOUNTPOINT")
	}
	
	mountpoint := flag.Arg(0)
		
	fmt.Printf("Reading config:%s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	fmt.Println("Generating filesystem")
	fs := filesystem.NewTriggerFS()
	for path, event := range config {
		if path[len(path)-1] == '/' {
			// directory
			path = strings.TrimRight(path, "/")
			attr := parser.ConfigToAttr(event[0], true)
			fs.Add(path, event[0].Pattern, event[0].Exec, attr)
			fmt.Printf("Add dir: %s\n", path)
			// multiple entries possible
			for i := 1; i < len(event); i++ {
				attr := parser.ConfigToAttr(event[i], false)
				fmt.Printf("Add dir rule: %s\n", event[i].Pattern)
				fs.Add(path, event[i].Pattern, event[i].Exec, attr)
			}
		} else {
			// file
			// only one entry per file definition allowed
			attr := parser.ConfigToAttr(event[0], false)
			fmt.Printf("Add file: %s\n", path)
			fs.Add(path, event[0].Pattern, event[0].Exec, attr)
		}
	}
	
	fmt.Printf("%v\n",fs)
	fmt.Printf("Mounting on %s\n", mountpoint)
	nfs := pathfs.NewPathNodeFs(fs, &pathfs.PathNodeFsOptions{ClientInodes: true})
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount failed: %v\n", err)
	}
	
	fmt.Println("Filesystem ready")
	server.Serve()
}
