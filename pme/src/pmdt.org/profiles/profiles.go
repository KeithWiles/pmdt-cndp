// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package profiles

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	tlog "pmdt.org/ttylog"
	"strings"
)

// EventInfo per event in a profile
type EventInfo struct {
	EventName string `json:"EventName"`
	Display   string `json:"Display"`
	GroupType string `json:"GroupType"`
}

// Profile - My Event list
type Profile struct {
	Name        string      `json:"Name"`
	Description string      `json:"Description"`
	Events      []EventInfo `json:"Events"`
}

// Info - information on selected configuration
type Info struct {
	filename string              // filename used for the profiles
	Names    []string            // Slice of all profile names
	Profiles map[string]*Profile // Full data for each profile
}

var profiles Info

func init() {
	profiles.Profiles = make(map[string]*Profile)
}

const (
	// defaultFilename - Default name of the PMDT Config file
	defaultFilename = "pme_events"
)

// Parse - Read the list of Profiles from the JSON file
func Parse(file string) error {

	if len(file) == 0 {
		file = defaultFilename
	}

	if !strings.Contains(file, ".json") {
		file = fmt.Sprintf("%s.json", file)
	}
	profiles.filename = file

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("File error %s", err)
	}

	dec := json.NewDecoder(strings.NewReader(string(dat)))

	_, err = dec.Token()
	if err != nil {
		return fmt.Errorf("Decode Start of JSON file error %s", err)
	}

	profiles.Profiles = make(map[string]*Profile, 0)
	for dec.More() {
		var e Profile

		err := dec.Decode(&e)
		if err != nil {
			fmt.Printf("Decode failed %s\n", err)
			return fmt.Errorf("Decode JSON failed")
		}

		profiles.Profiles[e.Name] = &e
		profiles.Names = append(profiles.Names, e.Name)
	}

	_, err = dec.Token()
	if err != nil {
		return fmt.Errorf("Decode End of JSON file error %s", err)
	}

	return nil
}

// ByName - find and retrive the events by name
func ByName(name string) (eventStr, displayStr string, err error) {

	if len(profiles.Profiles) == 0 {
		return "", "", fmt.Errorf("profiles are not loaded")
	}

	ec, ok := profiles.Profiles[name]
	if !ok {
		return "", "", fmt.Errorf("%s: not found", name)
	}

	for _, e := range ec.Events {
		if len(eventStr) == 0 {
			eventStr += e.GroupType + e.EventName
			displayStr = e.Display
		} else {
			if e.GroupType == "{" {
				eventStr += ";" + e.GroupType + e.EventName
			} else {
				eventStr += ";" + e.EventName
			}
			if e.GroupType == "}" {
				eventStr += e.GroupType
			}
			displayStr += ";" + e.Display
		}
	}

	return eventStr, displayStr, nil
}

// DisplayString - return the display string for the event
func DisplayString(name string) (string, error) {

	tlog.DebugPrintf("Profile name: %s\n", name)

	_, str, err := ByName(name)
	if err != nil {
		return "", err
	}
	if len(str) == 0 {
		return "", fmt.Errorf("display string is empty")
	}

	return str, nil
}

// EventString - return just the EventStr value
func EventString(name string) (string, error) {

	str, _, err := ByName(name)
	if err != nil {
		return "", err
	}
	if len(str) == 0 {
		return "", fmt.Errorf("event string is empty")
	}

	return str, nil
}

// SetProfile to be used for other calls
func SetProfile(name string) error {

	_, ok := profiles.Profiles[name]
	if !ok {
		return fmt.Errorf("profile named %s not found", name)
	}

	return nil
}

// Profiles of profile data
func Profiles() map[string]*Profile {

	return profiles.Profiles
}

// NameStrings is the list of profile names used as keys into profile data
func NameStrings() []string {

	return profiles.Names
}

// Names is the list of profile names used as keys into profile data
func Names() []interface{} {

	names := make([]interface{}, 0)

	for _, n := range profiles.Names {
		names = append(names, n)
	}
	return names
}
