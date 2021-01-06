package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

//DefaultSettings stores default values
var DefaultSettings map[string]string = map[string]string{
	"PORT":            "7000",
	"HOST":            "0.0.0.0",
	"TIMEOUT_SECONDS": "10",
}

//Env Environment global object
type Env struct {
	localSettings map[string]string
}

func (e *Env) getEnvFromOS(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return ""
	}
	return value
}

//loads the env file provided in settings into
func (e *Env) parseSettingsFile(filename string) {

	e.localSettings = make(map[string]string)

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Printf("File %s not found, using default settings\n", filename)
		return
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("File %s could not be read, using default settings\n", filename)
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			line = strings.Trim(line, " ")
			pair := strings.SplitN(line, "=", 1)

			if len(pair) > 2 {
				continue
			}

			e.localSettings[pair[0]] = pair[1]
		}
	}

	fmt.Printf("Settings : %x\n", e.localSettings)
}

func (e *Env) fromMapIfExist(key string) string {
	value, ok := DefaultSettings[key]
	if !ok {
		return ""
	}

	return value
}

func (e *Env) getEnv(name string) string {
	//check if env exists in local settings, else get it from os, else return default settings
	value, exists := e.localSettings[name]
	if !exists {
		value = e.getEnvFromOS(name)
		if value == "" {
			return e.fromMapIfExist(name)
		}

		return value
	}
	return value
}
