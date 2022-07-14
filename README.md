
# CadenceRadio
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kenellorando/cadence?style=flat-square)

**Cadence** is a fully-featured HTTP API application suite for *Icecast/Liquidsoap* web radios.

This project ships an API server, web-frontend, and built-in music-metadata database wrapper for custom-built stream service _Icecast_ and _Liquidsoap_ (also included out of the box). All components are containerized, with releases for amd64 and armv7.

See a demo of this on [https://cadenceradio.com/](https://cadenceradio.com/).

## Get Started

1. Configure the `cadence/config/cadence.env` file.
   1. Configure the `CSERVER_MUSIC_DIR` to an absolute path of a directory on your local system which you want to play music. The target is not recursively searched. The default location is `/music/`.
   2. (optional) Configure the rate limiter by setting the `CSERVER_REQRATELIMIT` to the number of seconds you wish to timeout users after they make song requests. Setting this value to `0` will disable rate limiting.
3. `docker compose up`

This will pull pre-built images (with working default configurations) for each of the services and start them with your `cadence.env` file options. The frontend interface is accessible by default at `localhost:8080`.

Running Cadence like this is perfectly fine for local usage, but if you plan to expose Cadence beyond your local network, you'll need to tweak a few configuration files so the radio services are password protected (see next section).

### Publicly Accessible Configuration (Password Protected)

> **Warning**: The way this repo currently handles configuration involves adding passwords to files which you may accidentally commit, so be careful.

1. Configure the `cadence_icecast2/config/cadence.xml` file.
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to the endpoint you expect your audience to connect to. This can be a DNS name, an IP address, or "localhost" (if you are running locally).
2. Configure the `cadence_liquidsoap/config/cadence.liq` file.
   1. Change all instances of `hackme` to a new password.
3. `docker compose up`

## Building for Development

If you are developing for Cadence and need to run exactly what you have without using pre-built images:

1. `docker compose up --build`
