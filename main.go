package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"time"
	"log"
	"fmt"
	"os"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

)

const VERSION string = "0.1"


func main() {
	
	//configfile := "config.json"
	var configfile string
	flag.StringVar(&configfile,"c", "config.conf", "Configfile")
	flag.StringVar(&configfile,"config", "config.conf", "Configfile")
	var fuseopts string
	flag.StringVar(&fuseopts,"fuseoptions", "", "Options for fuse")
	debug := flag.Bool("debug", false, "print fuse debugging messages")
	nosizecache := flag.Bool("nosizecache", false, "disable filesize caching")
	prebuildcache := flag.Bool("prebuildcache", false, "create sizecache by running all file exec commands once on startup")
	version := flag.Bool("version", false, "print version and exit")
	ttl := flag.Float64("ttl", 1.0, "attribute/entry cache TTL")
	gid := flag.Int("gid", os.Geteuid(), "set group id")
	uid := flag.Int("uid", os.Getgid(), "set user id")
	//loglevel := flag.Int("loglvl", 1, "set loglevel 1-3")
	flag.Parse()
	
	if *version {
		fmt.Printf("TriggerFS v%s\n", VERSION)
		return
	}
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  triggerfs MOUNTPOINT")
	}
	
	//logger := stdlog.GetFromFlags()
	
	mountpoint := flag.Arg(0)
		
	fmt.Printf("Reading config:%s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	fmt.Println("Generating filesystem")
	fs := filesystem.NewTriggerFS()
	
	// commandline args overwrite configfile options
	if *nosizecache {
		config.Caching = false
	}	
	if *prebuildcache {
		config.PrebuildCache = true
	}
	
	fs.BaseConf["triggerFS"] = config
	
	for path, cfg := range config.Dir {
		//log.Printf("Add dir: %s\n", path)
		attr := parser.ConfigToAttr(cfg, true)
		fs.AddDir(path, attr)
	}
	for path, cfg := range config.File {
		//log.Printf("Add file: %s\n", path)
		attr := parser.ConfigToAttr(cfg, false)
		fs.AddFile(path, cfg.Exec, attr)
	}
	for path, cfg := range config.Pattern {
		//log.Printf("Add pattern: %s\n", path)
		attr := parser.ConfigToAttr(cfg, false)
		fs.AddPattern(path, cfg.Exec, attr)
	}
	
	//spew.Dump(config)
	
	nfs := pathfs.NewPathNodeFs(fs, nil)
	
	fmt.Printf("Mounting on %s\n", mountpoint)
	opts := &nodefs.Options{
		AttrTimeout:  time.Duration(*ttl * float64(time.Second)),
		EntryTimeout: time.Duration(*ttl * float64(time.Second)),
		Debug:        *debug,
		LookupKnownChildren: false,
		Owner: &fuse.Owner{
			Uid: uint32(*uid),
			Gid: uint32(*gid),
		},
	}

	fmt.Printf("%v %v %v",opts,uid,gid)
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), opts)
	if err != nil {
		log.Fatalf("Mount failed: %v\n", err)
	}
	
	fmt.Println("Filesystem ready")
	server.Serve()
}
