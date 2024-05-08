## 📻 About

**Cadence** (or *CadenceRadio*) is an all-in-one internet radio suite. 

The project ships *Icecast* and *Liquidsoap* working out-of-the-box, made complete by a *Cadence API* providing song request, library search, album artwork, and real-time stream information in a browser UI.

**See a live demo on [cadenceradio.com](https://cadenceradio.com/)!**

<img src="https://user-images.githubusercontent.com/17265041/219263637-6971ce33-209a-4eb5-b67e-547f271dc3c8.png" width="600" >

## 🏃 Get Started

An interactive installation script is provided. Users familiar with Docker can be up and running in ~5 minutes.

### Server Preparation

- [Docker Engine](https://docs.docker.com/engine/install/) and [Docker Compose V2](https://docs.docker.com/compose/install/) are installed.
- You have some music files (e.g. `.mp3`, `.flac`) with title and artist metadata.

### Installation

Clone the Cadence repository to your server, then run the following:

```bash
$ chmod +x ./install.sh
$ ./install.sh
```

You will be prompted to provide an absolute path to a directory containing music, a stream hostname, a rate limit timeout, a service password, and optional reverse proxy configuration. If you need help figuring out what values to use, refer to the [Installation Guide](https://github.com/kenellorando/cadence/wiki/Installation#interactive-prompt-guide). 

Your radio stack will automatically launch and Cadence's web UI will become accessible at `localhost:8080`.

After initial installation, simply run `docker compose pull` to check for container updates, then `docker compose up` to start your station again. Run `./install.sh` again at any time to reconfigure. If you make changes to code locally, run `docker compose up --build` to build and run.

## 🔬 Technical Details

### Architecture
<details>
<summary><i>Basic Architecture</i></summary>

<img src="https://user-images.githubusercontent.com/17265041/228726513-e71775c4-dce4-4ef3-b4c2-1bbd37999769.png" width="800" >

</details>

If you're interested in implementation details, [this blog post](https://cuddle.fish/posts/2022-11-08-cadence) does a dive into how a basic *Icecast/Liquidsoap* web radio works and the value Cadence provides.

### API Reference for Custom Clients
Cadence's GitHub Wiki also hosts an [API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference) with complete request/response details, useful for anyone developing custom scripts or clients for their station.

### Discord Server Integration
Cadence installations can be directly integrated with Discord Servers using [CadenceBot](https://github.com/za419/CadenceBot). CadenceBot allows you to control your station through Discord chat and listen to the radio in voice channels! 
You can quickly demo a CadenceBot by [adding it to your Discord server](https://discord.com/api/oauth2/authorize?client_id=372999377569972224&permissions=274881252352&scope=bot).
