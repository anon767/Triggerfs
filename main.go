package main

import (
	"configurablefs/domain"
	"flag"
	"fmt"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"log"
)

func main() {

	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  configurablefs MOUNTPOINT FOLDER")
	}

	orig := flag.Arg(0)
	fmt.Printf("%s is mirrored\n", orig)
	nfs := pathfs.NewPathNodeFs(&domain.CFS{FileSystem: pathfs.NewLoopbackFileSystem(orig)}, nil)
	dest := flag.Arg(1)
	server, _, err := nodefs.MountRoot(dest, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	fmt.Printf("%s is mountpoint\n", dest)

	server.Serve()
}
