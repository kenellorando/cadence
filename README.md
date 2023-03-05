# CadenceRadio

**Cadence** is an all-in-one web radio suite that lets you start a self-hosted radio website.

In minutes, you can create an audio broadcast complete with library search, song request, album artwork, a browser UI, and real-time stream information working out-of-the-box.

All components are mostly pre-configured so there is hardly any configuration required to get started. Simply provide target directory containing music files, set a few service passwords and hostnames, and deploy!

**[See a live demo!](https://cadenceradio.com/)**

## 🖼️ Image Gallery
<details>
<summary>Browser UI Screenshot</summary>

![cadence5.1 browser ui](https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png)

</details>

<details>
<summary>Basic Service Architecture</summary>

![cadence5.3 architecture](https://user-images.githubusercontent.com/17265041/220829527-411f76ca-884f-4bf4-8b44-3afeaca158fa.png)

</details>

## 🏃 Get Started

### Requirements
1. You must have [Docker](https://docs.docker.com/engine/install/) installed. If you are on a Linux server, additionally install the [Compose plugin](https://docs.docker.com/compose/install/linux/).

### Installation
1. Edit `config/cadence.env`
   1. Change all instances of `hackme` to a new password.
   2. Set `CSERVER_MUSIC_DIR` to an absolute path which contains music files (`.mp3`, `.flac`) for play. The target is not recursively searched.
   3. Set `CSERVER_REQRATELIMIT` to an integer that sets the song request cooldown period in seconds. Set this value to `0` to disable rate limiting.
2. Edit `config/icecast.xml`
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to a URL you expect your audience to connect to. This value is what is set in the UI's stream source. This may be a DNS name or a public or private IP address. You can leave the default value `localhost` if your radio is meant to be accessible locally only.
3. Edit `config/liquidsoap.liq`
   1. Change all instances of `hackme` to a new password.
   2. If you changed `CSERVER_MUSIC_DIR` in step 1, change any instances of the default value `/music/` to match it here.
4. (_Optional_) Edit `config/nginx.conf`
   1. For advanced users deploying Cadence to a server with DNS, Cadence ships with a reverse proxy that can forward requests based on domain-name to backend services. Simply configure the `server_name` values with your domain names. The stream server domain should match the value you set in step 2.
5. `docker compose up`

After configuration is initially completed, you may simply run `docker compose up` again in the future to start your station.

### Accessing Services

- Assuming you kept the default values above, Cadence will become accessible in a browser at `localhost:8080`.
- If you optionally followed step 4, open firewall port `80` and point DNS to your server.

## 👩‍💻 Developing

### Building the Stack
If you changed code or updated a container image, append the `--build` flag to rebuild exactly what you have.

1. `docker compose down; docker compose up --build`

### Enable Development API
Cadence provides an optionally-enabled API with special administrative controls that may be useful for testing. Don't enable development mode on a production server.

1. Edit `config/cadence.env`.
   1. Set `CSERVER_DEVMODE` to `1` (enabled).

### API Reference
[Cadence's API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference) provides usage details and complete request/response examples.
