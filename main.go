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
	
	// configfile
	var configfile string
	flag.StringVar(&configfile,"c", "config.conf", "config file")
	flag.StringVar(&configfile,"config", "config.conf", "config file")
	
	// title
	var title string
	flag.StringVar(&configfile,"title", "", "set title of fs")
	
	// triggerFS options
	nosizecache := flag.Bool("nosizecache", false, "disable file size caching")
	prebuildcache := flag.Bool("prebuildcache", false, "create sizecache on startup")
	updatetree := flag.Bool("updatetree", false, "add files matching patters to fs tree after they've been accessed once")
	version := flag.Bool("version", false, "print version and exit")
	//loglevel := flag.Int("loglvl", 1, "set loglevel 1-3")
	
	//fuse options
	gid := flag.Int("gid", os.Geteuid(), "set group id")
	uid := flag.Int("uid", os.Getgid(), "set user id")
	debug := flag.Bool("debug", false, "print fuse debugging messages")
	ttl := flag.Float64("ttl", 1.0, "attribute/entry cache TTL")
	
	flag.Parse()
	
	if *version {
		fmt.Printf("TriggerFS v%s\n", VERSION)
		return
	}
	if len(flag.Args()) < 1 {
		fmt.Println("Usage:\n  triggerfs [<arg>] MOUNTPOINT\n")
		fmt.Println("Arguments:")
		flag.PrintDefaults()
		os.Exit(2)
		
	}
	
	//logger := stdlog.GetFromFlags()
	
	mountpoint := flag.Arg(0)
		
	fmt.Printf("Reading config: %s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	// commandline args overwrite configfile options
	if title != "" {
		config.Title = title
	}	
	if *nosizecache {
		config.Caching = false
	}	
	if *prebuildcache {
		config.PrebuildCache = true
	}	
	if *updatetree {
		config.UpdateTree = true
	}
	
	// set defaults
	if config.Title == "" {
		config.Title = "triggerfs"
	}
	
	
	// make fs and attach config
	fmt.Println("Generating filesystem")
	fs := filesystem.NewTriggerFS()
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
	
	fmt.Println(nfs.String())
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
	connector := nodefs.NewFileSystemConnector(nfs.Root(), opts)
	
	mountOpts := &fuse.MountOptions{
		AllowOther:    true,
		DisableXAttrs: true,
		Debug:         *debug,
		FsName:        "triggerFS",
		Name:          config.Title,
	}
	server, err := fuse.NewServer(connector.RawFS(), mountpoint, mountOpts)
	if err != nil {
		log.Fatalf("Mount failed: %v\n", err)
	}

	
	
	fmt.Println("Filesystem ready.")
	server.Serve()
}


