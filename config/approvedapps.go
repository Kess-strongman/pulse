package config

// Copyright 2017 Gary Barnett. All rights reserved.
// Use of this source code is governed by the same BSD-style
// license that the Golang Team uses.
//
// This implements a simple configuration file, which is stored as a json object
// the configuration is intended to support being updated at run-time with the
// option to save the configuration so it is reloaded on restart, or just to
// be applied for this runtime session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

// Config is a global containing the current configuration info
var AA Approvedapps

//Configuration holds the runtime config info
type Approvedapps struct {
	AppsArray []Approvedapp
	sync.RWMutex
}

type Approvedapp struct {
	Token       string
	Application string
}

func (a *Approvedapps) GetApplications() []string {
	a.RLock()
	defer a.RUnlock()
	appsArray := []string{}
	for _, key := range a.AppsArray {
		appsArray = append(appsArray, key.Application)
	}
	return appsArray
}

func (a *Approvedapps) FindTokens(checkkey string) bool {
	exists := false
	a.RLock()
	for _, key := range a.AppsArray {
		if key.Token == checkkey {
			exists = true
		}
	}
	defer a.RUnlock()
	return exists

}
func (a *Approvedapps) FindApplication(checkapp string) bool {
	exists := false
	a.RLock()
	for _, app := range a.AppsArray {
		if app.Application == checkapp {
			exists = true
		}
	}
	defer a.RUnlock()
	return exists
}

func (a *Approvedapps) GetTokenUsingAppid(appid string) string {
	a.RLock()
	for _, each := range a.AppsArray {
		fmt.Println(each)
		if each.Application == appid {
			return each.Token
		}
	}
	defer a.RUnlock()
	return " "
}
func (a *Approvedapps) SetTokenUsingAppid(appid string, newToken string) string {
	a.RLock()
	for _, each := range a.AppsArray {
		fmt.Println(each)
		if each.Application == appid {
			each.Token = newToken
			return "token changed"
		}
	}
	defer a.RUnlock()
	return "application not found"
}

func (a *Approvedapps) DeleteApplication(appid string) string {
	a.RLock()
	i := 0
	for i = 0; i < len(a.AppsArray); i++ {
		if a.AppsArray[i].Application == appid {
			break
		}
	}
	a.AppsArray = append(a.AppsArray[:i], a.AppsArray[i+1:]...)
	defer a.RUnlock()
	return " "
}
func (a *Approvedapps) AddApplication(newappid string, newToken string) string {
	a.RLock()
	var temp Approvedapp
	temp.Application = newappid
	temp.Token = newToken
	a.AppsArray = append(a.AppsArray, temp)
	defer a.RUnlock()
	return "application added"
}

// SaveToFile saves the configuration
func (a *Approvedapps) SaveToFile(fname string) {
	if fname == "" {
		fname = "assets/config1.json"
	}
	a.Lock()
	defer a.Unlock()
	b, err := json.Marshal(a)
	if err != nil {
		fmt.Println("Could not unmarshall ", err)
		return
	}
	err = ioutil.WriteFile(fname, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// GetJSON returns the JSON encoding of the configuration
func (a *Approvedapps) GetJSON() string {

	a.Lock()
	defer a.Unlock()
	b, err := json.Marshal(a)
	if err != nil {
		fmt.Println(err)
		return "could not convert config to JSON"
	}
	return string(b)
}

// LoadFromFile loads the configuration from a file
func (a *Approvedapps) LoadFromFile(fname string) error {
	if fname == "" {
		fname = "assets/approvedapps.json"
	}
	file, openerr := os.Open(fname)
	if openerr != nil {
		return openerr
	}
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&a.AppsArray)
	if err != nil {
		return err

	}
	return nil
}
