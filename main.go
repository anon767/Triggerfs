package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"fmt"

	"log"
	"path/filepath"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/davecgh/go-spew/spew"

)

func main() {
	
	//configfile := "config.json"
	var configfile string
	flag.StringVar(&configfile,"c", "config.conf", "Configfile")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  triggerfs MOUNTPOINT")
	}
	
	mountpoint := flag.Arg(0)
		
	fmt.Printf("Reading config:%s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	fmt.Println("Generating filesystem")
	fs := filesystem.NewTriggerFS()
	for path, cfg := range config.Dir {
		path = "/" + path
		fmt.Printf("Add dir: %s\n", path)
		attr := parser.ConfigToAttr(cfg, true)
		fs.Add(path, "", cfg.Exec, attr)
	}
	for path, cfg := range config.File {
		path = "/" + path
		fmt.Printf("Add file: %s\n", path)
		attr := parser.ConfigToAttr(cfg, false)
		fs.Add(path, "", cfg.Exec, attr)
	}
	for path, cfg := range config.Pattern {
		path = "/" + path
		dirpath, base := filepath.Split(path)
		fmt.Printf("Add pattern: %s into %s\n", base, dirpath)
		
		dircfg := cfg
		dircfg.Permission = "0755"
		dircfg.Size = 4096
		attr := parser.ConfigToAttr(dircfg, true)
		fs.Add(dirpath, "", "", attr)
		
		
		attr = parser.ConfigToAttr(cfg, false)
		fs.Add(dirpath, base, cfg.Exec, attr)
	}
		
		//if path[len(path)-1] == '/' {
			//// directory
			//path = strings.TrimRight(path, "/")
			//attr := parser.ConfigToAttr(event[0], true)
			//fs.Add(path, event[0].Pattern, event[0].Exec, attr)
			//fmt.Printf("Add dir: %s\n", path)
			//// multiple entries possible
			//for i := 1; i < len(event); i++ {
				//attr := parser.ConfigToAttr(event[i], false)
				//fmt.Printf("Add dir rule: %s\n", event[i].Pattern)
				//fs.Add(path, event[i].Pattern, event[i].Exec, attr)
			//}
		//} else {
			//// file
			//// only one entry per file definition allowed
			//attr := parser.ConfigToAttr(event[0], false)
			//fmt.Printf("Add file: %s\n", path)
			//fs.Add(path, event[0].Pattern, event[0].Exec, attr)
		//}
	//}
	
	
	spew.Dump(fs)
	spew.Dump(config)
	fmt.Printf("Mounting on %s\n", mountpoint)
	nfs := pathfs.NewPathNodeFs(fs, &pathfs.PathNodeFsOptions{ClientInodes: true})
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount failed: %v\n", err)
	}
	
	fmt.Println("Filesystem ready")
	server.Serve()
}
