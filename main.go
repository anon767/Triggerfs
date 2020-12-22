package main

import (
	"triggerfs/filesystem"
	"triggerfs/parser"
	"flag"
	"time"
	"log"
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
	flag.StringVar(&title,"title", "", "set title of fs")
	
	// triggerFS options
	sizecache := flag.Bool("sizecache", false, "enable file size caching")
	nosizecache := flag.Bool("nosizecache", false, "disable file size caching")
	
	prebuildcache := flag.Bool("prebuildcache", false, "create sizecache on startup")
	noprebuildcache := flag.Bool("noprebuildcache", false, "don't create sizecache on startup")
	
	updatetree := flag.Bool("updatetree", false, "add files matching patters to fs tree after they've been accessed once")
	noupdatetree := flag.Bool("noupdatetree", false, "don't add files matching patters to fs tree after they've been accessed once")
	
	version := flag.Bool("version", false, "print version and exit")
	loglevel := flag.Int("loglvl", 0, "set loglevel 0-3")
	
	//fuse options
	gid := flag.Int("gid", os.Geteuid(), "set group id")
	uid := flag.Int("uid", os.Getgid(), "set user id")
	debug := flag.Bool("debug", false, "print fuse debugging messages")
	ttl := flag.Float64("ttl", 1.0, "attribute/entry cache TTL")
	
	flag.Parse()
	
	log.SetOutput(os.Stdout)
	
	if *version {
		log.Printf("TriggerFS v%s\n", VERSION)
		return
	}
	if len(flag.Args()) < 1 {
		log.Println("Usage:\n  triggerfs [<arg>] MOUNTPOINT\n")
		log.Println("Arguments:")
		flag.PrintDefaults()
		os.Exit(2)
		
	}
	
	//logger := stdlog.GetFromFlags()
	
	mountpoint := flag.Arg(0)
		
	log.Printf("Starting TriggerFS v%s\n", VERSION)
	
	log.Printf("Reading config: %s\n", configfile)
	config := parser.Parseconfig(configfile)
	
	// commandline args overwrite configfile options
	if title != "" {
		config.Title = title
	}	
	if *loglevel > 0 {
		config.LogLevel = int(*loglevel)
	}	
	
	if *sizecache {
		config.Caching = true
	}
	if *nosizecache {
		config.Caching = false
	}
	
	if *prebuildcache {
		config.PrebuildCache = true
	}	
	if *noprebuildcache {
		config.PrebuildCache = false
	}	
	
	if *updatetree {
		config.UpdateTree = true
	}
	if *noupdatetree {
		config.UpdateTree = false
	}
	
	// set defaults
	if config.Title == "" {
		config.Title = "triggerfs"
	}
	
	
	// make fs and attach config
	log.Println("Generating filesystem")
	fs := filesystem.NewTriggerFS()
	fs.BaseConf["triggerFS"] = config
	fs.LogLevel = *loglevel
	
	for path, cfg := range config.Dir {
		attr := parser.ConfigToAttr(cfg, true)
		fs.AddDir(path, attr)
	}
	for path, cfg := range config.File {
		attr := parser.ConfigToAttr(cfg, false)
		fs.AddFile(path, cfg.Exec, attr)
	}
	for path, cfg := range config.Pattern {
		attr := parser.ConfigToAttr(cfg, false)
		fs.AddPattern(path, cfg.Exec, attr)
	}
	
	//spew.Dump(config)
	
	nfs := pathfs.NewPathNodeFs(fs, nil)
	
	log.Println(nfs.String())
	log.Printf("Mounting on %s\n", mountpoint)
	// set mount options
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
	
	// set fuse options
	mountOpts := &fuse.MountOptions{
		AllowOther:    true,
		DisableXAttrs: true,
		Debug:         *debug,
		FsName:        "triggerFS",
		Name:          config.Title,
	}
	server, err := fuse.NewServer(connector.RawFS(), mountpoint, mountOpts)
	if err != nil {
		log.Fatalf("ERROR Mount failed: %v\n", err)
	}

	
	
	log.Println("Filesystem ready.")
	server.Serve()
}


