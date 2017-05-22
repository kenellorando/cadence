# Cadence Radio [Listen Here](http://cadenceradio.com/)
## Background
Cadence is an online radio heavily inspired by 4chan/8ch's [R/a/dio](http://r-a-d.io/). The project is my first endeavour to practice a full-range of IT-skills. It currently utilizes technologies in front-end design with HTML and CSS, site scripting with JavaScript, and some networking and server management with both Windows Server 2016 and Debian. 

Currently, back-end work for a queryable music database, to handle music metadata, is now being built with PHP and SQL (MySQL). The database is populated using a [populator tool](https://github.com/kenellorando/cadence-database-populator) programmed in Python.

Once the database is completed, the final main feature will be a song request bot.

## Acknowledgements
### Technical Assistance
* [Ryan Hodin](https://github.com/za419) (HTML, CSS, JS)
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
