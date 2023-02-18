# CadenceRadio

**Cadence** is an all-in-one web radio suite, allowing you to start a self-hosted internet radio website in minutes.

The project ships with _Icecast_ and _Liquidsoap_ built-in, complemented by a _Cadence_ API server providing music search, song request, artwork, a UI, and real-time stream information. 

Cadence ships all components mostly pre-configured with each each other so there is hardly any work needed to get started. Set a target directory containing your music files, set a few service passwords and hostnames, and you're done! The Cadence stack can be deployed in a single command.

**Try the live demo! [cadenceradio.com](https://cadenceradio.com/)**

## 🖼️ Image Gallery
<details>
<summary>Cadence Browser UI</summary>

![cadence5.1 browser ui](https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png)

</details>

<details>
<summary>Cadence Architecture</summary>

![cadence5 architecture](https://user-images.githubusercontent.com/17265041/185465196-66fc2249-e43a-46f7-a12f-dbde9aaf8172.png)

</details>

## 🏃 Get Started

### Requirements
1. You must have Docker installed. If you are on a Linux server, install the [Compose plugin](https://docs.docker.com/compose/install/linux/).

### Installation
1. Edit `cadence/config/cadence.env`.
   1. Set `CSERVER_MUSIC_DIR` to an absolute path of a directory on your system which contains music files (`.mp3`, `.flac`) to play. The target is not recursively searched. The default value is `/music/`.
   2. Set `CSERVER_REQRATELIMIT` to an integer value that represents a song request cooldown period in seconds. Set this value to `0` to disable request rate limiting. The default value is `180`.
2. Edit `cadence_icecast2/config/cadence.xml`.
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to a URL you expect your audience to connect to. Cadence uses this value to set the stream source in the UI. This may be a DNS name or a public or internal IP address. You can leave the default value `localhost` if your radio is meant to be accessible locally only.
3. Edit `cadence_liquidsoap/config/cadence.liq`:
   1. Change all instances of `hackme` to a new password.
   2. If you changed the `CSERVER_MUSIC_DIR` value in step 1, change any instances of the default value `/music/` to match it here.
4. `docker compose up`
   1. On older versions of Docker, use `docker-compose up` instead.

### Accessing Services
Assuming no changes were made to port numbers or the hostnames in the steps above:

- The UI is accessible in a browser at `localhost:8080`
- API server requests may also be sent to the `localhost:8080` path. See the API Reference for more details.
- The audio stream is accessible at `localhost:8000/cadence1`.

## 👩‍💻 Developing

### Building the Stack
If you are developing and need to rebuild exactly what you have, use the `--build` flag.

1. `docker compose down; docker compose up --build`

### Development Mode
Cadence provides special administrative controls that may be useful for testing in an optional development API. As the name implies, don't enable development mode on a production server. 

1. Edit `cadence/config/cadence.env`.
   1. Set `CSERVER_DEVMODE` from `0` (disabled) to `1` (enabled).

### API Reference
If you are developing a custom client for Cadence Radio stacks, [Cadence's Wiki API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference) provides usage details and complete request/response examples.
