package main

import (
	"configurablefs/filesystem"
	"flag"
	"fmt"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"log"
)

func main() {

	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  configurablefs ROOT MOUNTPOINT")
	}

	destinationRoot := flag.Arg(0)
	fmt.Printf("%s is mirrored\n", destinationRoot)
	nfs := pathfs.NewPathNodeFs(&filesystem.CFS{FileSystem: pathfs.NewLoopbackFileSystem(destinationRoot)}, nil)
	mountpoint := flag.Arg(1)
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	fmt.Printf("%s is mountpoint\n", mountpoint)

	server.Serve()
}
