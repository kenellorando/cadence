package main

func main() {
	// Grab initialization values from the yaml config
	c := initConfig()
	// Start the logger using the config defined log level
	initLogger(c.LogLevel)
}
