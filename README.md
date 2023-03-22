<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/17265041/226766588-30a7bb71-490b-4182-a6b0-b22409f6ec3b.png" width="300">
    <source media="(prefers-color-scheme: light)" srcset="https://user-images.githubusercontent.com/17265041/226766586-1a55554e-33a4-4fb2-8259-49359f3f3525.png" width="300">
    <img alt="Cadence logo.">
  </picture>
  <p align="center"><b>All-in-one, self-hosted web radio suite.</b></p>
  <div align="center"><img align="center" src="https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png" height="300" ></div>
</p>

## 📻 About

**Cadence** (or *CadenceRadio*) is an all-in-one internet radio suite with library search, song request, album artwork, and real-time stream information in a browser UI. Simply run an interactive installation script, provide some music, and enjoy!

**See a live demo on [cadenceradio.com](https://cadenceradio.com/)!**

## 🏃 Start Here

### Requirements
- You have [Docker](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/) installed.

### Installation
```bash
chmod +x ./install.sh
./install.sh
```

You will be prompted to provide a music directory path, a stream hostname, a rate limit timeout, a service password, and optional DNS. Your radio stack will automatically launch and Cadence's web UI will become accessible at `localhost:8080`.

After initial installation, simply run `docker compose up` to start your station. Run `./install.sh` again at any time to reconfigure.

## 🦔 Documentation

- [Cadence: Self-Hosted Web Radio Suite](https://kenellorando.notion.site/Cadence-Self-Hosted-Web-Radio-Suite-d1f0184b5eeb4882a3d6f78d582b2de6)
- [API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference)
- [Installation Guide](https://github.com/kenellorando/cadence/wiki/Installation)
