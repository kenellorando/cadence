## 📻 About

**Cadence** (or *CadenceRadio*) is an all-in-one internet radio suite. 

The project ships *Liquidsoap* working out-of-the-box, made complete by a *Cadence API* which receives the audio stream, serves it to listeners, and provides song request, library search, album artwork, and real-time stream information in a browser UI.

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
$ ./install.sh
```

You will be prompted to provide an absolute path to a directory containing music, a rate limit timeout, a service password, and optional reverse proxy configuration. If you need help figuring out what values to use, refer to the [Installation Guide](https://github.com/kenellorando/cadence/wiki/Installation#interactive-prompt-guide). 

Your radio stack will start in the background and Cadence's web UI will become accessible at `localhost:8080`. The stream is served from that same address, so there is no second host or port to configure.

Your music library is read in the background after startup. A large library takes a while, and the search tab says so until it has finished.

After initial installation:

```bash
$ docker compose up -d          # start the station
$ docker compose logs -f        # follow the logs
$ docker compose down           # stop it
$ docker compose pull           # check for container updates
$ docker compose up -d --build  # rebuild after local code changes
```

Run `./install.sh` again at any time to reconfigure. To start over completely, `docker compose down -v` also removes the database volume; the next start rebuilds the library index from your music directory.

### Serving over HTTPS

The web UI and the audio stream share one origin, so a single certificate for your domain covers both. `config/nginx.conf.example` carries the TLS server block to add and a certbot command to obtain a certificate. Serving the page over HTTPS while the audio stays on plain HTTP would be blocked by the browser as mixed content, which is why they are not split across two hosts.

## 🔬 Technical Details

### Architecture
<details>
<summary><i>Basic Architecture</i></summary>

*Note: this diagram predates Cadence serving the stream itself, and still shows Icecast as a separate service. Liquidsoap now sends its output to Cadence, which fans it out to listeners.*


<img src="https://user-images.githubusercontent.com/17265041/228726513-e71775c4-dce4-4ef3-b4c2-1bbd37999769.png" width="800" >

</details>

If you're interested in implementation details, [this blog post](https://cuddle.fish/posts/2022-11-08-cadence) covers how a basic *Icecast/Liquidsoap* web radio works and the value Cadence provides. Cadence has since taken over what Icecast was doing -- it receives the stream from Liquidsoap and serves listeners directly -- but the post remains a good explanation of the shape of the problem.

### API Reference for Custom Clients
Cadence's GitHub Wiki also hosts an [API Reference](https://github.com/kenellorando/cadence/wiki/API-Reference) with complete request/response details, useful for anyone developing custom scripts or clients for their station.

### Discord Server Integration
Cadence installations can be directly integrated with Discord Servers using [CadenceBot](https://github.com/za419/CadenceBot). CadenceBot allows you to control your station through Discord chat and listen to the radio in voice channels! 
You can quickly demo a CadenceBot by [adding it to your Discord server](https://discord.com/api/oauth2/authorize?client_id=372999377569972224&permissions=274881252352&scope=bot).
