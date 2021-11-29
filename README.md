# Cadence ♥
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kenellorando/cadence?style=flat-square)
![GitHub issues](https://img.shields.io/github/issues/kenellorando/cadence?style=flat-square)

Cadence is an original web radio originally started as a personal project back when I was a student in February 2017. Development continues to this day. The current iteration (4.x) is a Go webserver which exposes an API for interaction with a Postgres database, Liquidsoap music client, and Icecast stream server. The live API provides users with the ability to search a music library and even make song requests. Cadence is not built with any frameworks. Check it out on [cadenceradio.com](https://cadenceradio.com).

## Starting
For API functionality, Cadence requires a connection with any Postgres 12 database, local or remote, plus a connection to a Liquidsoap music client and Icecast stream server, which I recommend are run locally. Cadence comes with a start script `START.sh` which interactively prompts for **all backend** necessary configurations, like connection addresses and target directories, making setup for API development easy. This script stores all configuration data in the system environment. An empty database will be automatically configured and populated by the webserver on startup. 

Beyond using the API, I do not recommend you try setting up a full working instance with my frontend. This is an old project, and Cadence's frontend was not designed with abstraction in mind. Even if you set up the backend and all stream services perfectly, the frontend _as is_ will not be compatible. Because I haven't abstracted configurations out of the frontend, it will attempt to connect to official `cadence1` music streams and the `cadenceradio.com` API. To hook the Cadence webserver into your own stream server, you will also need to do a significantly time consuming replacement/rewrite of [player.js](https://github.com/kenellorando/cadence/blob/master/public/js/player.js) file and the [index.html](https://github.com/kenellorando/cadence/blob/master/public/index.html#L45), replacing stream variables with the particulars of your own.

To start the server, run:
```
$ ./START.sh
```
Alternatively, if the configurations are already set, you may forego the configuration prompts by simply running:
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
