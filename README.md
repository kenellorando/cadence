
# CadenceRadio
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kenellorando/cadence?style=flat-square)

**Cadence** is a fully-featured HTTP API application suite for *Icecast/Liquidsoap* web radios.

Out of the box, this project ships an API server, web-frontend, and autopopulating music-metadata database with custom _Icecast_ and _Liquidsoap_ containers. A Compose file will let you set up an entire radio stack in minutes. All components have releases for amd64 and armv7. 

See a demo of this on [https://cadenceradio.com/](https://cadenceradio.com/).

## Get Started

### Running Locally

1. Edit the `cadence/config/cadence.env` file:
   1. Set the `CSERVER_MUSIC_DIR` value to an absolute path of a directory on your local system which you want to play music. The target is not recursively searched. The default location is `/music/`.
   2. Set the `CSERVER_REQRATELIMIT` to the number of seconds you wish to timeout users after they make song requests. Setting this value to `0` will disable rate limiting.
3. `docker compose up`

Running `docker compose up` will start all of the Cadence services. The frontend interface is accessible by default at `localhost:8080`.

### Password Protecting Services

If you plan to expose Cadence beyond your local network, you'll need to tweak a few configuration files so the radio services are password protected (see next section).

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
