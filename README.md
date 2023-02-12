# CadenceRadio

**Cadence** is an all-in-one web radio suite, allowing you to start a self-hosted internet radio website in minutes.

The project ships with _Icecast_ and _Liquidsoap_ built-in, complemented by a _Cadence_ API server providing music search, song request, artwork, a UI, and real-time stream information. 

Cadence ships all components mostly pre-configured with each each other so there is hardly any work required to get started. Set a target directory containing your music, set a few service passwords and hostnames, and you're all set! The Cadence stack can be deployed in a single command.

**See a live demo on [cadenceradio.com](https://cadenceradio.com/)!**

<details>
<summary>Cadence Browser UI</summary>

![image](https://user-images.githubusercontent.com/17265041/214758418-8002cd64-2e3b-4a17-9104-1ad29d20ded6.png)

</details>

<details>
<summary>Cadence Architecture</summary>

![cadence5 architecture](https://user-images.githubusercontent.com/17265041/185465196-66fc2249-e43a-46f7-a12f-dbde9aaf8172.png)

</details>

## 🏃 Get Started

### Requirements
1. You must have Docker installed. If you are running Docker on a Linux server, additionally install the [Compose plugin](https://docs.docker.com/compose/install/linux/).

### Installation
1. Edit `cadence/config/cadence.env`.
   1. Set `CSERVER_MUSIC_DIR` to an absolute path of a directory on your system which contains your music files to play. The target is not recursively searched.
   2. Set `CSERVER_REQRATELIMIT` to an integer value of seconds to timeout users after they make song requests. Set this value to `0` to disable rate limiting.
2. Edit `cadence_icecast2/config/cadence.xml`.
   1. Change all instances of `hackme` to a new password.
   2. Set the `<hostname>` value to a URL you expect your audience to connect to. Cadence uses this value to set the stream source in the UI. This may be a DNS name, a public or internal IP address, or default to `localhost` if the radio is meant to be accessible from the host machine only.
3. Edit `cadence_liquidsoap/config/cadence.liq`:
   1. Change all instances of `hackme` to a new password.
   2. If you changed the `CSERVER_MUSIC_DIR` value in step 1, change any instances of the default value `"/music/"` to match it here.
4. `docker compose up`
   1. On older versions of Docker, use `docker-compose up` instead.

### Accessing Services
Assuming no changes were made to port numbers or the hostnames in the steps above:

- The UI is accessible in a browser at `localhost:8080`
- API server requests may also be sent to the `localhost:8080` path. See the API Reference for more details.
- The audio stream is accessible at `localhost:8000/cadence1`.

## 👩‍💻 Development

### Enabling _Development Mode_ and Building the Stack Locally
Cadence provides an optional secret API, DevMode, that allow special administrative control that may be useful for testing. See the API Reference for development commands. As the name implies, don't enable DevMode on a production server. 

1. Edit `cadence/config/cadence.env`.
   1. Set `CSERVER_DEVMODE` from `0` (disabled) to `1` (enabled).
2. `docker compose up --build`

### API Reference
See [Cadence's GitHub Wiki for API Documentation](https://github.com/kenellorando/cadence/wiki/API-Reference) for complete details and request/response examples.

### Kubernetes Deployments
It is possible to deploy a Cadence stack to a Kubernetes cluster. Manifests and additional information are provided in [cadence-k8s](https://github.com/kenellorando/cadence-k8s).
