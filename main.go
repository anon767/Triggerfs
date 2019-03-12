package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"fmt"
	"time"
	"log"
	
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

)

func main() {
	
	//configfile := "config.json"
	var configfile string
	flag.StringVar(&configfile,"c", "config.conf", "Configfile")
	var fuseopts string
	flag.StringVar(&fuseopts,"fuseoptions", "", "Options for fuse")
	debug := flag.Bool("debug", false, "print debugging messages.")
	//version := flag.Bool("version", false, "print version.")
	ttl := flag.Float64("ttl", 1.0, "attribute/entry cache TTL.")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  triggerfs MOUNTPOINT")
	}
	
	mountpoint := flag.Arg(0)
		
	fmt.Printf("Reading config:%s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	fmt.Println("Generating filesystem")
	fs := filesystem.NewTriggerFS()
	fs.BaseConf["triggerFS"] = config
	
	for path, cfg := range config.Dir {
		//fmt.Printf("Add dir: %s\n", path)
		attr := parser.ConfigToAttr(cfg, true)
		fs.AddDir(path, attr)
	}
	for path, cfg := range config.File {
		//fmt.Printf("Add file: %s\n", path)
		attr := parser.ConfigToAttr(cfg, false)
		fs.AddFile(path, cfg.Exec, attr)
	}
	for path, cfg := range config.Pattern {
		//fmt.Printf("Add pattern: %s\n", path)
		attr := parser.ConfigToAttr(cfg, false)
		fs.AddPattern(path, cfg.Exec, attr)
	}
	
	//spew.Dump(config)
	
	nfs := pathfs.NewPathNodeFs(fs, &pathfs.PathNodeFsOptions{ClientInodes: true})
	
	fmt.Printf("Mounting on %s\n", mountpoint)
	opts := &nodefs.Options{
		AttrTimeout:  time.Duration(*ttl * float64(time.Second)),
		EntryTimeout: time.Duration(*ttl * float64(time.Second)),
		Debug:        *debug,
	}
	
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), opts)
	if err != nil {
		log.Fatalf("Mount failed: %v\n", err)
	}
	
	fmt.Println("Filesystem ready")
	server.Serve()
}
