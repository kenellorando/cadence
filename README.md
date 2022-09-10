# CadenceRadio
**Cadence** is a fully-featured HTTP API web radio software suite.

It ships everything you need to start a web radio station, including an API server (with search, request, now playing, artwork, and stream information functions), an interactive browser UI, an autopopulating music-metadata database, and custom built-in _Icecast_ and _Liquidsoap_ integration.

![cadence5](https://user-images.githubusercontent.com/17265041/189464889-b6a67b78-8d9d-4aef-a142-2494448f26a4.JPG)

Cadence is essentially a metadata broker between your music files and stream services, with a programmatic interface for your audiences. The project ships all components pre-configured to work with each each other so there is hardly any configuration to do.

To configure Cadence, set a target directory containing your music, set a few service passwords and hostnames, and you're good to go! Cadence has container releases for amd64 and armv7. All components are containerized, and a Compose file sets the entire radio stack up in minutes. See a live demo on [cadenceradio.com](https://cadenceradio.com/)!

![cadence5 architecture](https://user-images.githubusercontent.com/17265041/185465196-66fc2249-e43a-46f7-a12f-dbde9aaf8172.png)



## Get Started

> **Warning**: The way this repo currently handles configuration involves adding passwords to files which you may accidentally commit, so be careful.

1. Edit the `cadence/config/cadence.env` file:
   1. Set the `CSERVER_MUSIC_DIR` value to an absolute path of a directory on your local system which you want to play music. The target is not recursively searched. The default location is `/music/`.
   2. Set the `CSERVER_REQRATELIMIT` to the number of seconds you wish to timeout users after they make song requests. Setting this value to `0` will disable rate limiting.
2. Configure the `cadence_icecast2/config/cadence.xml` file.
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to the endpoint you expect your audience to connect to. Cadence uses this value to set the stream source in the UI. This can be a DNS name, an IP address, or default `localhost` (for running locally).
3. Configure the `cadence_liquidsoap/config/cadence.liq` file.
   1. Change all instances of `hackme` to a new password.
4. `docker compose up`

Running `docker compose up` will start all Cadence services. The UI is accessible by default at `localhost:8080`.

## Building for Development

If you are developing and need to rebuild exactly what you have locally:

1. `docker compose up --build`

## Container Repositories
- https://hub.docker.com/repository/docker/kenellorando/cadence
- https://hub.docker.com/repository/docker/kenellorando/cadence_icecast2
- https://hub.docker.com/repository/docker/kenellorando/cadence_liquidsoap

## API Quick Reference
> See the `API` tab on https://cadenceradio.com for more details and request/response examples.

`POST /api/search` 
- Search the server's music library. Cadence is additionally aware of audio file metadata for album, year, and genre, meaning you may also try to search by any of those parameters.

`POST /api/request/id` 
- Submit a request for a specific song to be played. This call accepts a song ID, which is an impermanent label Cadence applies to each song and is not an inherent attribute of the song file itself. Therefore, you should execute a search for an ID as a prerequisite to request if you want a guaranteed result. Also consider using /api/request/bestmatch to request the number one search result automatically.

`POST /api/request/bestmatch` 
- This call executes a search, selects the number one result (the most relevant result to the search), and submits a request for it. This is Cadence's equivalent to Google's "I'm Feeling Lucky".

`GET /api/nowplaying/metadata` 
- Gets all text metadata on the currently playing song. If there are multiple audio files on the server which share the exact same title and artist, Cadence will return the first result only. Does not return album art. To get album art, use /api/nowplaying/albumart

`GET /api/nowplaying/albumart` 
- Gets the base64-encoded album art of the currently playing song.

`GET /api/listeners` 
- Returns the number of connected listeners.

`GET /api/listenurl` 
- Returns the direct audio listen URL. This is used for programmatically setting any audio sources.

`GET /api/version` 
- Returns the server version.

`GET /api/radiodata/sse` 
- This is a special server-sent event (SSE) connection API. Clients may connect to this endpoint, and as long as they stay connected, will receive server-pushed data as it changes live on the backend without the need for the client to poll. The data is sent within one second after a change is observed. All data transfer happens through only one event source and is differentiated by unique events. At the moment, this event source will monitor and notify changes for title, artist, listenurl, and listeners. All of these can also be fetched on-demand through their respective API endpoints.
