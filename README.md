# Cadence Radio (http://www.kenellorando.com)
## Background
Cadence is an online radio heavily inspired by 4chan/8ch's [R/a/dio](http://r-a-d.io/). The project is my first endeavour to practice a full-range of IT-skills. It currently utilizes technologies in front-end design with HTML and CSS, site scripting with JavaScript, and some networking and server management using Windows Server 2016. Back-end work for a music database is also currently being built using the MySQL RDBMS. The database is filled with thousands of non-redundant entries of music metadata using a populator tool I programmed in Java.

Future plans for the radio include a front-end database query feature (PHP) and an advanced back-end request bot (C#).

## Acknowledgements
### Technical Assistance
* [Ryan Hodin](https://github.com/za419) (HTML, CSS, JS)
* [Bobby Ton](https://github.com/bobbyt1997) (Java)
### User Testers
* Michael Folk
* Jakob Frank

## Contributing
To contribute to Cadence Radio, first install [Git LFS](https://git-lfs.github.com/), as our Space Station theme uses files hosted on LFS.

As `git clone` only permits serial file download, LFS suggests disabling LFS for the clone operation and then pulling the files separately, as this can be done in parallel. The commands suggested are (taken from the [LFS Tutorial](https://github.com/git-lfs/git-lfs/wiki/Tutorial)):

    GIT_LFS_SKIP_SMUDGE=1 git clone https://github.com/username/my_lfs_repo.git destination_dir
    #git lfs ls-files # optionally see all the - showing the lfs files are not checked out
    git lfs pull
    #git lfs ls-files # optionally see all the * showing the lfs files are checked out
