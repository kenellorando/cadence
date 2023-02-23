# CadenceRadio

**Cadence** is an all-in-one web radio suite that you can use to start a self-hosted radio website in minutes.

The project ships with _Icecast_ and _Liquidsoap_ built-in, complemented by a _Cadence_ API server providing music search, song request, artwork, a UI, and real-time stream information. 

Cadence ships all components mostly pre-configured with each each other so there is hardly any configuration required. Simply set a target directory containing your music files, set a few service passwords and hostnames, and deploy!

**[Try the live demo!](https://cadenceradio.com/)**

## 🖼️ Image Gallery
<details>
<summary>Cadence Browser UI</summary>

![cadence5.1 browser ui](https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png)

</details>

<details>
<summary>Cadence Architecture</summary>

![cadence5.3 architecture](https://user-images.githubusercontent.com/17265041/220829527-411f76ca-884f-4bf4-8b44-3afeaca158fa.png)

</details>

## 🏃 Get Started

### Requirements
1. You must have Docker installed. If you are on a Linux server, install the [Compose plugin](https://docs.docker.com/compose/install/linux/).

### Installation
1. Fork or clone the repository.
2. Edit `config/cadence.env`
   1. Change all instances of `hackme` to a new password.
   2. Set `CSERVER_MUSIC_DIR` to an absolute path which contains music files (`.mp3`, `.flac`) for play. The target is not recursively searched. The default value is `/music/`.
   3. Set `CSERVER_REQRATELIMIT` to an integer that sets the song request cooldown period in seconds. Set this value to `0` to disable rate limiting. The default value is `180`.
3. Edit `config/icecast.xml`
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to a URL you expect your audience to connect to. This value is what is set in the UI's stream source. This may be a DNS name or a public or private IP address. You can leave the default value `localhost` if your radio is meant to be accessible locally only.
4. Edit `config/liquidsoap.liq`
   1. Change all instances of `hackme` to a new password.
   2. If you changed `CSERVER_MUSIC_DIR` in step 1, change any instances of the default value `/music/` to match it here.
5. (_Optional_) Edit `config/nginx.conf`
   1. For advanced users deploying Cadence to a server with DNS, Cadence ships with a reverse proxy which will forward requests based on domain-name to backend services. Simply configure the `server_name` values with your domain names.
6. `docker compose up`

### Accessing Services

- Assuming you kept the default values above, Cadence will become accessible in a browser at `localhost:8080`.
- If you optionally followed step 5 to make Cadence publicly accessible, open firewall port `80` and point DNS to your server.

## 👩‍💻 Developing

### Building the Stack
If you changed code or updated a container image, and need to rebuild exactly what you have, use the `--build` flag.

`docker compose down; docker compose up --build`

### Enable Development API
Cadence provides special administrative controls that may be useful for testing though an optionally enabled API. Don't enable development mode on a production server. See the API reference for more details.

1. Edit `config/cadence.env`.
   1. Set `CSERVER_DEVMODE` to `1` (enabled).

### API Reference
Interested in developing custom scripts or clients for your Cadence Radio? [Cadence's API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference) provides usage details and complete request/response examples.
