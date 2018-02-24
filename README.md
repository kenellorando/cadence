# Cadence Radio ♥ [http://cadenceradio.com](http://cadenceradio.com/)
## About
Cadence is an online radio heavily inspired by [R/a/dio](http://r-a-d.io/). Originally started in February 2017, the project was my first endeavour to practice a full range of IT-skills including front-end design, back-end programming, databases, Linux server management, and cybersecurity. Cadence is served on a Node.js webserver interacting with a Mongo music database, both configured around a Liquidsoap/Icecast stream server. 

The server typically plays all the genres of music I like, mostly a mix of classic rock, synthpop, and metal. I occasionally override the server and play a certain genre, artist, or the same song on an endless loop.

**Cadence Radio is a DMCA compliant, "non-commercial educational" (NCE) broadcast. As an NCE, Cadence Radio is non-profit and does not accept advertisements for its webpages or broadcasts.**

## Discord Bot
Add Cadence Radio to your Discord server! Created by [Ryan Hodin](https://github.com/za419)
* [Add to server](https://discordapp.com/oauth2/authorize?client_id=372999377569972224&scope=bot&permissions=1)
* [Source](https://github.com/za419/CadenceBot)

## Features
* A 24/7 audio livestream
* Song requests
* Over a dozen gorgeous themes in seasonal rotation

## Contributors
* [Ryan Hodin](https://github.com/za419) (Programming, Design, Media)
* [Bobby Ton](https://github.com/bobbyt1997) (Design)
* [Jakob Frank](https://github.com/jakobfrank) (Media)
* Michael Farrell (Security)
* Mike Folk (QA)
* Zheng Guo (Translations)
* Karen Santos (Design)
* Kelvin Chang (Design)

## Contributing
To contribute to Cadence Radio, first install [Git LFS](https://git-lfs.github.com/), as our Space Station theme uses files hosted on LFS.

As `git clone` only permits serial file download, LFS suggests disabling LFS for the clone operation and then pulling the files separately, as this can be done in parallel. The commands suggested are (taken from the [LFS Tutorial](https://github.com/git-lfs/git-lfs/wiki/Tutorial)):

    GIT_LFS_SKIP_SMUDGE=1 git clone https://github.com/username/my_lfs_repo.git destination_dir
    #git lfs ls-files # optionally see all the - showing the lfs files are not checked out
    git lfs pull
    #git lfs ls-files # optionally see all the * showing the lfs files are checked out
