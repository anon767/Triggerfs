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

type event struct {
	Permission string `json:"permission"`
	Pattern    string `json:"pattern"`
	Exec       string `json:"exec"`
}

func Parseconfig(configfile string) (events []event) {
	jsonFile, err := os.Open(configfile)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &events)
	return events
}

func ExecFile(file string) string {
	out, err := exec.Command(file).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func MatchFile(file string, event []event) (string, bool) {
	for i := 0; i < len(event); i++ {
		r, _ := regexp.Compile(event[i].Pattern)
		if r.MatchString(file) {
			return event[i].Exec, true
		}
	}
	return "", false
}
