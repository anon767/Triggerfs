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
	Permission string `json:"permission"`
	Pattern    string `json:"pattern"`
	Exec       string `json:"exec"`
}

func Parseconfig(configfile string) (events []Event) {
	jsonFile, err := os.Open(configfile)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &events)

	return events
}

func (event Event) ExecFile(file string) string {
	out, err := exec.Command(event.Exec, file).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (event Event) MatchFile(file string) bool {
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
	return Event{"", "", ""}, false
}
