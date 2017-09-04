# Cadence Radio [cadence.kenellorando.com](http://cadence.kenellorando.com/)
## About
Cadence is an online radio heavily inspired by 4chan/8ch's [R/a/dio](http://r-a-d.io/). The project is my first endeavour to practice a full-range of IT-skills. The code housed in this repository is primarily front end HTML, CSS, and JavaScript. 

By default, the server plays all the genres of music I like, mostly classic rock, synthpop, and metal. But you'll also find anything from rap to orchestral to Jpop. I occasionally override the server and manually DJ.

Though the primary feature development of the radio has ended, I'm continually improving existing code as I learn. I intend to implement bigger features like a queryable database and a request bot once I've comfortably gained some new skills.

**Cadence Radio is a DMCA compliant, non-commercial webcast made for educational purposes.**

## Features
* A 24/7 audio livestream
* Automatically updating song info display
* Three permanent live-background themes plus more seasonally rotating ones!

## Acknowledgements
### Technical Assistance
* [Ryan Hodin](https://github.com/za419) (HTML, CSS, JavaScript)
* [Bobby Ton](https://github.com/bobbyt1997) (Java)
### User Testers
* Michael Folk
* Jakob Frank

## Contributing
To contribute to Cadence Radio, first install [Git LFS](https://git-lfs.github.com/), as the now-removed Space Station theme used files hosted on LFS, and we may use LFS again in the future.

As `git clone` only permits serial file download, LFS suggests disabling LFS for the clone operation and then pulling the files separately, as this can be done in parallel. The commands suggested are (taken from the [LFS Tutorial](https://github.com/git-lfs/git-lfs/wiki/Tutorial)):

    GIT_LFS_SKIP_SMUDGE=1 git clone https://github.com/username/my_lfs_repo.git destination_dir
    #git lfs ls-files # optionally see all the - showing the lfs files are not checked out
    git lfs pull
    #git lfs ls-files # optionally see all the * showing the lfs files are checked out

## Todo
* Repopulate database with current server tracks.
* Copy database from music server to web server.
