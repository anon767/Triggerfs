package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type Event struct {
	Path       string `json:"path"` // maybe drop Pattern?
	Permission string `json:"permission"`
	Pattern    string `json:"pattern"`
	Exec       string `json:"exec"`
}

func Parseconfig(configfile string) (events []Event) {
	//var events Event
	jsonFile, err := os.Open(configfile)
	if err != nil {
		fmt.Println("open:", err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("read:", err)
	}

	err = json.Unmarshal(byteValue, &events)
	if err != nil {
		fmt.Println("unmarshall error:", err)
	}

	return events
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

func EventsMatchFile(file string, events []Event) (Event, bool) {
	for i := 0; i < len(events); i++ {
		if events[i].MatchFile(file) {
			return events[i], true
		}
	}
	return Event{"", "", "", ""}, false
}
