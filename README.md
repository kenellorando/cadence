# Cadence Radio [http://cadenceradio.com](http://cadenceradio.com/)
## About
Cadence is an online radio heavily inspired by [R/a/dio](http://r-a-d.io/). Originally started in February 2017, the project is my first endeavour to practice a full range of IT-skills. Development continues today with a focus on back-end technology.

The server typically plays all the genres of music I like, a mix of classic rock, synthpop, and metal. I occasionally override the server and play a certain genre, artist, or the same song on an endless loop.

**Cadence Radio is a DMCA compliant, non-commercial webcast made for educational purposes.**

## Features
* A 24/7 audio livestream
* Automatically updating song info display
* Over six gorgeous themes in seasonal rotation

## Contributors
* [Ryan Hodin](https://github.com/za419) (Programming, Design, Media)
* [Bobby Ton](https://github.com/bobbyt1997) (Design)
* [Jakob Frank](https://github.com/jakobfrank) (Media)
* Michael Farrell (Security)

## Contributing
To contribute to Cadence Radio, first install [Git LFS](https://git-lfs.github.com/), as the now-removed Space Station theme used files hosted on LFS, and we may use LFS again in the future.

As `git clone` only permits serial file download, LFS suggests disabling LFS for the clone operation and then pulling the files separately, as this can be done in parallel. The commands suggested are (taken from the [LFS Tutorial](https://github.com/git-lfs/git-lfs/wiki/Tutorial)):

    GIT_LFS_SKIP_SMUDGE=1 git clone https://github.com/username/my_lfs_repo.git destination_dir
    #git lfs ls-files # optionally see all the - showing the lfs files are not checked out
    git lfs pull
    #git lfs ls-files # optionally see all the * showing the lfs files are checked out
