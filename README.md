# Cadence ♥
Cadence is a web radio originally started as a personal project back when I was a student in February 2017. Its goal: practice a full range of tech skills, from development to server administration. Development continues to this day.

The application is a Go webserver which exposes an API for interaction with a database and Icecast stream server. The website acts as an interface to this API, providing users with the ability to search the music library and even make song requests. Cadence is not built with any frameworks.

[cadenceradio.com](http://cadenceradio.com)

## Stack
- Go
- Postgres

## Starting
Cadence comes with a start script which interactively prompts for configurations, like addresses and directories, before starting the server. All configuration data is stored in the system environment. The database is automatically configured and populated by the webserver on startup.
```
$ ./START.sh
```
Alternatively, if the configurations are already set, you may forego the configuration step.
```
$ go run server/*.go
```

## Contributors
* [Ryan Hodin](https://github.com/za419)
* [Bobby Ton](https://github.com/bobbyt1997)
* [Jakob Frank](https://github.com/jakobfrank)
* Mike Farrell
* Mike Folk
* Zheng Guo
* Karen Santos
* Kelvin Chang
