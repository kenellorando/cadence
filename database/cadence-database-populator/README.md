# Cadence Database Populator
Made for [Cadence Radio](https://github.com/kenellorando/cadence), this tool collects all relevant metadata from a given directory of audio files. This metadata includes song titles, artist names, album names, song lengths, and absolute paths. This data is then inserted into Cadence's local database.

![Cadence Database Populator success](https://raw.githubusercontent.com/kenellorando/cadence-database-populator/master/sample-output.jpg)

### Run instructions:
- Copy the repository to the Cadence database server. 
- Ensure Python 3.x+ is installed and set the path if necessary
- Install the MySQL-Python connector
```
# From the /mysql-connector-python-2.0.4/ directory run 
$ python setup.py install
```

- Edit the configuation at the top of populator.py file if necessary
```
# Path of directory holding music
path = "C:/Path/To/Music"

# Database configuration stored here
config = {
    'user': 'populator',
    'password': 'populator1',
    'host': '127.0.0.1',
    'database': 'cadence',
}
```

- Run the populator with
```
$ python populator.py
```

### Project is dependent on:
- [Python/MySQL Connector](https://pypi.python.org/pypi/mysql-connector-python/2.0.4)
- [TinyTag](https://github.com/devsnd/tinytag)
