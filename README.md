# Cadence ♥
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kenellorando/cadence?style=flat-square)
![GitHub issues](https://img.shields.io/github/issues/kenellorando/cadence?style=flat-square)

## About 

Cadence is a web radio with a heavy focus on anime music. Streaming 24/7, Cadence provides interaction with a song metadata database and open-source stream software Icecast. As this allows users the ability to search a song library and make song requests, Cadence effectively simulates a song request taking virtual DJ.

Over two thousand lines of code written from scratch compose this application. In the current iteration (4.x) of Cadence, no frontend framework is employed and the backend is written in Go. Originally hosted from a dorm room homelab, all Cadence services now run on AWS.

This project was originally started while I was a student in February 2017 as an endeavour to practice a full range of tech skills, from development to server administration. In 2019, the experience I gained building Cadence were directly responsible for landing my first job as a DevOps Engineer. Today, I continue to run Cadence and make occasional improvements. For questions and comments, please feel to reach out on the [Cadence Github repository](https://github.com/kenellorando/cadence).

## Starting Locally
- If setting up database functionality, a connection with any Postgres 12 database, local or remote is required. 
- If setting up radio functionality, a Liquidsoap music client and Icecast stream server are required. Due to a limitation of the Liquidsoap request system, Liquidsoap must be able to receive telnet connections from the Cadence application. Finally, a filesystem of audio files must be available to both Cadence and Liquidsoap. For these reasons, I highly recommend Cadence and all stream services run together on the same local machine.
- If setting up frontend compatibility, please note this repository is hardcoded with official (the real `cadenceradio.com` API) Icecast addresses, so setting up _as is_ will not work out of the box with a local installation. This is an old project, and Cadence's frontend was not originally designed with abstraction in mind. To hook the Cadence webserver into your own stream server, you will also need to do a time consuming but not impossible rewrite of [player.js](https://github.com/kenellorando/cadence/blob/master/public/js/player.js) file and the [index.html](https://github.com/kenellorando/cadence/blob/master/public/index.html#L45), replacing stream variables with the particulars of your own.

Optional: Run `git submodule update --init --recursive` to update the `public/static` directory with image and video media.

Cadence comes with a start script `START.sh` which interactively prompts for **all backend** necessary configurations like connection addresses and target directories, making setup a little bit easier provided relevant supporting services are available. This script stores all configuration data as shell variables. An empty database will be automatically configured and populated by the webserver on startup. 

```
$ ./START.sh
```
Alternatively, if the configurations are already set, you may forego the configuration prompts and start the backend server as configurations live as environment variables.
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
