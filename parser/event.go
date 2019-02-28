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
	
	//spew.Dump(config)

	return config
}

//func (entry Entry) MatchFile(file string) bool { //here maybe check if file==entry.Path?
	//matched, err := regexp.MatchString(entry.Pattern, file)
	//if err != nil {
		//log.Fatal(err)
	//}
	//if matched {
		//return true
	//}
	//return false
//}

//func EntrysMatchFile(file string, config []Config) (Entry, bool) {
	////for i := 0; i < len(config); i++ {
		////if config[i][file].MatchFile(file) {
			////return entrys[i], true
		////}
	////}
	//return Entry{"", "", "", 0, 0, 0, 0}, false
//}


func ConfigToAttr(config Entry, dir bool) (*fuse.Attr, uint32) {
	attr := &fuse.Attr{}
	permission := uint32(0644)
	mode := uint32(fuse.S_IFREG)
	
	if dir {
		permission = 0755
		mode = fuse.S_IFDIR
	}
	
	if config.Permission != "" {
		//int_permission, err := strconv.Atoi(config.Permission)
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

	
	return attr, permission
}
