# CadenceRadio

**Cadence** is an all-in-one suite that lets you start a self-hosted web radio website.

In minutes, create an internet broadcast with library search, song request, album artwork, and real-time stream information in a browser UI. All components are mostly pre-configured to work out-of-the-box. Simply run an interactive installation script and deploy!

**[See a live demo!](https://cadenceradio.com/)**

## 🖼️ Gallery
<details>
<summary>Browser UI Screenshot</summary>

![cadence5.1 browser ui](https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png)

</details>

<details>
<summary>Basic Service Architecture</summary>

![cadence5.3 architecture](https://user-images.githubusercontent.com/17265041/220829527-411f76ca-884f-4bf4-8b44-3afeaca158fa.png)

</details>

## 🏃 Start Here

### Requirements
You must have [Docker](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/) installed.

### Installation
```bash
chmod +x ./install.sh
./install.sh
```

You will be prompted for a music directory path, a stream hostname, a rate limit timeout, a service password, and optional DNS. After the last prompt, the radio stack will automatically launch and Cadence's web UI will become accessible at `localhost:8080`. 

After initial installation, simply run `docker compose up` to start your station. Use `install.sh` again at any time to reconfigure inputs.

## 📚 Knowledge Base
Cadence's GitHub Wiki provides various resources to help you use, administrate, and build clients for your station.

- [API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference)
- [Development and Code Style Guide](https://github.com/kenellorando/cadence/wiki/Development-and-Code-Style)
- [Installation Guide](https://github.com/kenellorando/cadence/wiki/Installation)
