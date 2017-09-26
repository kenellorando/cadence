const PORT = 8080;
const IP = 'localhost';
const DB_URL = 'mongodb://localhost:27017/cadence'

var express = require('express');
var path = require('path');
var bodyParser = require('body-parser');
var MongoClient = require('mongodb').MongoClient;

var mm = require('musicmetadata');

var app = express();


// Parse incoming data
app.use(bodyParser.urlencoded({
  extended: true
}));

// Point to publicly served files
app.use(express.static(path.join(__dirname, 'public')));

// Connect to database. Populate.
MongoClient.connect(DB_URL, function (err, db) {
  if (err) {
    return console.log(err);
  }

  var fs = require('fs');
  var walkPath = '/home/ken/Music';
  var walkPath = './test';

  var walk = function (dir, done) {
    fs.readdir(dir, function (error, list) {
      if (error) {
        return done(error);
      }

      var i = 0;
      (function next() {
        var file = list[i++];

        if (!file) {
          return done(null);
        }

        file = dir + '/' + file;
        fs.stat(file, function (error, stat) {
          if (stat && stat.isDirectory()) {
            walk(file, function (error) {
              next();
            });
          } else {
            var parser = mm(fs.createReadStream(file), function (err, metadata) {
              if (err) {
                throw err;
              }
              //console.log(metadata);
              var songInfo = '{ "title":"' + metadata.title + '", "artist":"' + metadata.artist + '", "album":"' + metadata.album + '", "path":"' + file + '"}';
              console.log(typeof(songInfo));
              var songInfoObject = JSON.parse(songInfo);
              console.log(typeof(songInfoObject));
              //console.log(songInfo);
            });

            //console.log(file);
            next();
          }
        });
      })();
    });
  };

  // optional command line params
  //      source for walk path
  process.argv.forEach(function (val, index, array) {
    if (val.indexOf('source') !== -1) {
      walkPath = val.split('=')[1];
    }
  });

  console.log('-------------------------------------------------------------');
  console.log('processing...');
  console.log('-------------------------------------------------------------');

  walk(walkPath, function (error) {
    if (error) {
      throw error;
    } else {
      console.log('-------------------------------------------------------------');
      console.log('finished.');
      console.log('-------------------------------------------------------------');
    }
  });

  console.log("Database up to date.");
  db.close();
});


// Search, directed from aria.js AJAX
app.post('/search', function (req, res) {
  console.log("Received: " + JSON.stringify(req.body));
  // Web Server Console:
  // Received: {"search":"railgun"}
  console.log(req.body.search);

  // Database search
  MongoClient.connect(DB_URL, function (err, db) {
    if (err) {
      return console.log(err);
    }

    // TODO: Search DB

    console.log("Connection established to database.");
    db.close();
  });

  res.send("OK from ARIA!");
  res.end();
});

var server = app.listen(PORT, IP);