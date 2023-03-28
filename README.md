## 📻 About

<img align="right" src="https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png" width="450" >

**Cadence** (or *CadenceRadio*) is an all-in-one internet radio suite. 

The project ships with *Icecast* and *Liquidsoap* working out-of-the-box, complete with library search, song request, album artwork, and real-time stream information in a browser UI. Simply run an interactive installation script, provide some music, and enjoy!

**See a live demo on [cadenceradio.com](https://cadenceradio.com/)!**

## 🏃 Start Here

### Requirements
- You have [Docker](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/) installed.

### Installation
```bash
chmod +x ./install.sh
./install.sh
```

You will be prompted to provide a music directory path, a stream hostname, a rate limit timeout, a service password, and optional DNS. Your radio stack will automatically launch and Cadence's web UI will become accessible at `localhost:8080`. That's all there is to it!

After initial installation, simply run `docker compose up` to start your station. Run `./install.sh` again at any time to reconfigure. 

If you can't run a shell script for any reason, or if the scripted install was too easy for you, alternative installation methods including fully-manual configuration and even Kubernetes deployment steps are provided on the [Installation Guide](https://github.com/kenellorando/cadence/wiki/Installation).

## 🦔 Resources

If you're interested in implementation details, [Cadence: Self-Hosted Web Radio Suite](https://kenellorando.notion.site/Cadence-Self-Hosted-Web-Radio-Suite-d1f0184b5eeb4882a3d6f78d582b2de6) does a dive into how a typical web radio works and the value Cadence provides.

Cadence's GitHub Wiki also hosts an [API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference) with complete request/response details, useful for anyone developing custom scripts or clients for their station.
