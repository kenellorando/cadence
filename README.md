# Cadence Radio ♥ [http://cadenceradio.com](http://cadenceradio.com/)
## About
Cadence is an online radio heavily inspired by [R/a/dio](http://r-a-d.io/). Originally started in February 2017, the project is my first endeavour to practice a full range of IT-skills including front-end design, web programming, databases, and Linux server management. Development continues today with a focus on back-end technology.

The server typically plays all the genres of music I like, mostly a mix of classic rock, synthpop, and metal. I occasionally override the server and play a certain genre, artist, or the same song on an endless loop.

**Cadence Radio is a DMCA compliant, "non-commercial educational" (NCE) broadcast. As an NCE, Cadence Radio is non-profit and does not accept advertisements for its webpages or broadcasts.**

## Discord Bot
Add Cadence Radio to your Discord server! Created by [Ryan Hodin](https://github.com/za419)
* [Add to server](https://discordapp.com/oauth2/authorize?client_id=372999377569972224&scope=bot&permissions=1) 
* [Source](https://github.com/za419/CadenceBot)

## Features
* A 24/7 audio livestream
* Song requests
* Automatically updating song info display
* Over six gorgeous themes in seasonal rotation

## Contributors
* [Ryan Hodin](https://github.com/za419) (Programming, Design, Media)
* [Bobby Ton](https://github.com/bobbyt1997) (Design)
* [Jakob Frank](https://github.com/jakobfrank) (Media)
* Michael Farrell (Security)
* Mike Folk (QA)

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
