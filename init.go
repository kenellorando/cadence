package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/kenellorando/clog"
	"gopkg.in/yaml.v2"
)

const configFile = "CONFIG.yaml"

type Config struct {
	LogLevel int `yaml:"LOG_LEVEL"`
}

// Read from the yaml configuration file for initialization values
func initConfig() Config {
	var c Config
	source, err := ioutil.ReadFile(configFile)
	// Panic without the logger if config unreadable
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &c)
	if err != nil {
		panic(err)
	}
	return c
}

func initLogger(l int) {
	fmt.Print(l)
	// Initialize logging level
	logLevel := clog.Init(l)
	clog.Debug("initLogger", "Logging service initialized to level "+strconv.Itoa(logLevel)+".")
}
