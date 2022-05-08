![GitHub release (latest by date)](https://img.shields.io/github/v/release/kenellorando/cadence?style=flat-square)
![GitHub issues](https://img.shields.io/github/issues/kenellorando/cadence?style=flat-square)

# CadenceRadio

**Cadence** is a fully-featured web radio suite application. 

Cadence provides an API server, web-frontend, and music-metadata service for stream services _Icecast_ and _Liquidsoap_. With Cadence, users may search for music and make song requests through the browser.

As of version `4c.x` all software is fully containerized, meaning you are minutes away from running your own web radio. Check it out in action: [https://cadenceradio.com/](https://cadenceradio.com/).

---

For questions and comments, you are welcome to [open a discussion](https://github.com/kenellorando/cadence/discussions) on Github.


## Installation

### Using Docker Compose

1. Create a `/music` directory on your system root populated with audio files to play (if you want to use a different location, you can override the volume mounts in `docker-compose.yml`).
2. Change all service passwords. All default password values are `hackme`.
   1. `icecast2/config/icecast.xml`
   2. `liquidsoap/config/liquidsoap.liq` (this password must match the source password set in `icecast.xml`)
   3. Set the `CSERVER_DB_PASS` in `cadence/config/cadence.env`.
3. `docker-compose up`

That's it. Cadence's web interface will be available at `localhost:8080`. Icecast web will be available at `localhost:8000`, and the default stream mountpoint will play on `localhost:8000/cadence1`.

## Discord Bot

_[CadenceBot](https://github.com/za419/CadenceBot/issues)_, maintained by [Ryan Hodin](https://github.com/za419), is an configurable Discord interface to Cadence API servers. The bot accepts commands through Discord text channels, and relays music into voice channels!

## Contributors

![GitHub Contributors Image](https://contrib.rocks/image?repo=kenellorando/cadence)
