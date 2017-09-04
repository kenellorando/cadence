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
To contribute to Cadence Radio, please note that git submodules are used in the 
project. This means that in order to properly clone Cadence, you should pass 
`--recursive` to `git clone`, or alternatively you should run these two commands 
after cloning:

1. `git submodule init`

2. `git submodule update`

After these are complete, or after a clone with `--recursive`, submodules will be 
properly set up.

When working with Cadence, you should occasionally run `git submodule update 
--remote` to update the submodules.

If you set the configuration setting `status.submodulesummary`, ie if you run `git 
config status.submodulesummary 1`, then git will generate a short summary of 
changes to submodules when running commands like `status`. Additionally, `git 
diff` will provide some information about changes in submodules if passed 
`--submodule`.

The changelog generated is ignored, and should not be committed into the 
repository - Since it's basically just a styled version of the git log, the 
correct changelog for any commit can be generated simply by checking out that 
commit and running the generator. Thus, it should not be version controlled.

Because the changelog file isn't version controlled, you will not have a 
changelog file when you clone Cadence. If you need one, simply navigate to the 
`changelog` directory and run `generator.sh`: It does assume the current working 
directory is where it is stored. Wait for it to finish: It will have generated a 
file public_html/changelog.html, which is the changelog file for the current 
commit.

If you deploy a mirror of Cadence, this changelog should be kept up-to-date, and 
so the generator should be run every time new commits are added.

## Todo
* Repopulate database with current server tracks.
* Copy database from music server to web server.
