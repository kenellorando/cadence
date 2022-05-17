![GitHub release (latest by date)](https://img.shields.io/github/v/release/kenellorando/cadence?style=flat-square)
![GitHub issues](https://img.shields.io/github/issues/kenellorando/cadence?style=flat-square)

# CadenceRadio

**Cadence** is a fully-featured web radio application suite. 

Cadence is an API server, web-frontend, and built-in music-metadata database bundled with custom-built stream service containers _Icecast_ and _Liquidsoap_. 

Stream music over the internet and allow users to search for songs and make requests through the browser. You are minutes away from starting your own web radio. Check it out in action: [https://cadenceradio.com/](https://cadenceradio.com/).

For questions and comments, you are invited to [open a discussion](https://github.com/kenellorando/cadence/discussions) on Github.


## Get Started

### Running the Stack

1. `docker compose up`

Running the above will pull pre-built images (with working default configurations) for each of the services and start them. Running Cadence like this is perfectly fine for local usage, but if you plan to expose Cadence beyond your local machine, you'll need to tweak a few configuration files.

### Running the Stack (Custom Config)

> **Warning**: The way this repo handles configuration is early and primitive. Building locally involves adding passwords to files which you may accidentally commit, so be careful.

1. Configure the `cadence/config/cadence.env` file.
   1. Configure the `CSERVER_MUSIC_DIR` to an absolute path of a directory on your local system which you want to play music. The default location is `/music/`.
   2. (optional) Configure the rate limiter by setting the `CSERVER_REQRATELIMIT` to the number of seconds you wish to timeout users after they make song requests. Setting this value to `0` will disable rate limiting.
2. Configure the `cadence_icecast2/config/cadence.xml` file.
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to the endpoint you expect your audience to connect to. This can be a DNS name, an IP address, or "localhost" (if you are running locally).
3. Configure the `cadence_liquidsoap/config/cadence.liq` file.
   1. Change all instances of `hackme` to a new password.
4. `docker compose up`

### Building the Stack

If you are developing for Cadence and need to build exactly what you have without pulling pre-built images:

1. `docker compose up --build`


## Discord Bot

_[CadenceBot](https://github.com/za419/CadenceBot/issues)_, maintained by [Ryan Hodin](https://github.com/za419), is an configurable Discord interface to Cadence API servers. The bot accepts commands through Discord text channels, and relays music into voice channels!

## Contributors

![GitHub Contributors Image](https://contrib.rocks/image?repo=kenellorando/cadence)
