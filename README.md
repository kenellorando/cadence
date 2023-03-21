<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/17265041/226766588-30a7bb71-490b-4182-a6b0-b22409f6ec3b.png" width="400">
    <source media="(prefers-color-scheme: light)" srcset="https://user-images.githubusercontent.com/17265041/226766586-1a55554e-33a4-4fb2-8259-49359f3f3525.png" width="400">
    <img alt="Cadence logo.">
  </picture>
</p>

**Cadence** (or *CadenceRadio*) is an all-in-one suite that lets you start a self-hosted web radio website.

In minutes, create an internet broadcast with library search, song request, album artwork, and real-time stream information in a browser UI. All components are mostly pre-configured to work out-of-the-box. Simply run an interactive installation script, provide some music, and deploy!

**See a live deployment on [cadenceradio.com](https://cadenceradio.com/)!**

## 🖼️ Preview Gallery
<details>
<summary>Browser UI</summary>

![cadence5.1 browser ui](https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png)

</details>

<details>
<summary>Basic Service Architecture</summary>

![cadence5.3 architecture](https://user-images.githubusercontent.com/17265041/220829527-411f76ca-884f-4bf4-8b44-3afeaca158fa.png)

</details>

## 🏃 Start Here

### Requirements
- You must have [Docker](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/) installed.
- You have some familiarity self-hosting services on Linux.

### Installation
```bash
chmod +x ./install.sh
./install.sh
```

You will be prompted for a music directory path, a stream hostname, a rate limit timeout, a service password, and optional DNS. Your radio stack will automatically launch and Cadence's web UI will become accessible at `localhost:8080`.

After initial installation, simply run `docker compose up` to start your station. Use `install.sh` again at any time to reconfigure inputs.

## 🦔 Documentation

- [Cadence: Self-Hosted Web Radio Suite](https://kenellorando.notion.site/Cadence-Self-Hosted-Web-Radio-Suite-d1f0184b5eeb4882a3d6f78d582b2de6)
- [API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference)
- [Installation Guide](https://github.com/kenellorando/cadence/wiki/Installation)
