# Cadence Radio ♥ [http://cadenceradio.com](http://cadenceradio.com/)
## About
Cadence is an anime-inspired online radio. Originally started in February 2017, the project was my first endeavour to practice a full range of IT-skills in web server programming, front-end design, Linux administration, databases, networking, and cybersecurity.

Cadence is essentially a web application which interacts with an Icecast stream server. Cadence further provides users with the ability to search through a music database as well as make requests.

The music library is diverse, with songs inclusive of almost every genre. I occasionally override the server and play a certain genre, artist, or the same song on an endless loop.

**Cadence Radio is a DMCA compliant, "non-commercial educational" (NCE) broadcast. As an NCE, Cadence Radio is non-profit and does not accept advertisements for its webpages or broadcasts.**

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
