# CadenceRadio

**Cadence** is an all-in-one web radio suite that lets you start a self-hosted radio website.

In minutes, you can create an internet audio broadcast complete with library search, song request, album artwork, and a browser UI with real-time stream information working out-of-the-box.

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
1. Run `./install.sh`.
   1. You will be prompted to provide inputs: a music directory path, a stream hostname, a rate limit timeout, a service password, and optional DNS. Cadence should automatically start.

After initial installation, you may simply run `docker compose up` again in the future to start your station. Run `install.sh` again at any time to reconfigure.

### Accessing Services

- By default, Cadence will become accessible in a browser at `localhost:8080`.
- If you optionally provided DNS, open firewall port `80` and point DNS to your server.

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
