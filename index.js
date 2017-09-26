const PORT = 8080;
const IP = 'localhost';
const DB_URL = 'mongodb://localhost:27017/cadence';
const MUSIC_DIR = './test';


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

  // Create a music collection.
  db.createCollection("music", function (err, res) {
    if (err) {
      throw err;
    }
  })

  // Walk the directory
  var fs = require('fs');
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
              /*
              // Create a song object
              var songInfoString = '{ "title":"' + metadata.title + '", "artist":"' + metadata.artist + '", "album":"' + metadata.album + '", "path":"' + file + '"}';
              var songInfoObject = JSON.parse(songInfoString);
              */
              // Insert the object to the database
              db.collection("music").update({
                path: file
              }, {
                $set: {
                  'title': metadata.title,
                  "artist": metadata.artist,
                  "album": metadata.album
                }
              }, {
                upsert: true
              })
              if (err) {
                throw err;
              }
            })
          };
          next();
        });
      })();
    });
  };

  // optional command line params
  //      source for walk path
  process.argv.forEach(function (val, index, array) {
    if (val.indexOf('source') !== -1) {
      MUSIC_DIR = val.split('=')[1];
    }
  });

  walk(MUSIC_DIR, function (error) {
    if (error) {
      throw error;
    }
  });

  
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