package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"os/exec"
	"regexp"
	
)

type Event struct {
	Path       string `json:"path"`
	Permission string `json:"permission"`
	Pattern    string `json:"pattern"`
	Exec       string `json:"exec"`
}

type Events []Event

type Config map[string]Events

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

func (event Event) ExecCmd(filename string) string {
	// ignore filename, it can be a parameter or something in the future
	out, err := exec.Command("sh", "-c", event.Exec).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (event Event) MatchFile(file string) bool { //here maybe check if file==event.Path?
	matched, err := regexp.MatchString(event.Pattern, file)
	if err != nil {
		log.Fatal(err)
	}
	if matched {
		return true
	}
	return false
}

func EventsMatchFile(file string, config []Config) (Event, bool) {
	//for i := 0; i < len(config); i++ {
		//if config[i][file].MatchFile(file) {
			//return events[i], true
		//}
	//}
	return Event{"", "", "", ""}, false
}
