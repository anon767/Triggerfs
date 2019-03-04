package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"strconv" 
	"github.com/hanwen/go-fuse/fuse"
)

type Entry struct {
	//Path       string `json:"path"`
	Permission string `json:"permission"`
	Pattern    string `json:"pattern"`
	Exec       string `json:"exec"`
	Mtime      int `json:"mtime"`
	Ctime      int `json:"ctime"`
	Atime      int `json:"atime"`
	Size       int `json:"size"`
}

type Config map[string][]Entry

func Parseconfig(configFile string) (config Config) {
	
	byteValue, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("read config:", err)
	}
	
	dec := json.NewDecoder(strings.NewReader(string(byteValue)))
	for {
		if err := dec.Decode(&config); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		
	}

	return config
}


func ConfigToAttr(config Entry, dir bool) (*fuse.Attr) {
	attr := &fuse.Attr{}
	permission := uint32(0644)
	mode := uint32(fuse.S_IFREG)
	
	if dir {
		permission = 0755
		mode = fuse.S_IFDIR
	}
	
	if config.Permission != "" {
		int_permission, err := strconv.ParseUint(config.Permission, 8, 32)
		if err == nil {
			permission = uint32(int_permission)
		}
		
	}
	attr.Mode = mode | permission

	attr.Size  = uint64(config.Size)
	attr.Mtime = uint64(config.Mtime)
	attr.Atime = uint64(config.Atime)
	attr.Ctime = uint64(config.Ctime)

	
	return attr
}
