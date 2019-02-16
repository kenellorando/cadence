package main

func main() {
	c, db := getConfig()
	initLogger(c.LogLevel)
	initDatabase(db)
}
