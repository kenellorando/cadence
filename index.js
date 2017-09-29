const PORT = 8080;
const IP = '198.37.25.185';
const DB_URL = 'mongodb://localhost:27017/cadence';
const MUSIC_DIR = '/home/ken/Music';


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

  // Drop the music collection
  // db.collection("music").drop();

  // Rebuild the music collection.
  db.createCollection("music", function (err, res) {
    if (err) {
      // Drop and recreate
      db.collection("music").drop();
      db.createCollection("music");

      /*
      Currently, these do not fire from here 
      and must be typed manually through 
      the database console

      // Drop old indexes
      db.collection("music").dropIndexes();
      // Enable text searching
      db.executeDbAdminCommand({
        setParameter: 1,
        textSearchEnabled: true
      });
      // Set search index
      db.collection("music").createIndex({
        title: "text",
        artist: "text",
        album: "text"
      });
      */
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
        
        var extensions = [ // All recognized music extensions
            ".mp3",
            ".m4a",
            ".flac",
            ".ogg"
        ];
        var music=false;
        for (var i=0; i<extensions.length; ++i) {
            if (file.endsWith(extensions[i])) {
                music=true;
                break;
            }
        }
        if (music)
            return next();
        
        file = dir + '/' + file;
        fs.stat(file, function (error, stat) {
          if (stat && stat.isDirectory()) {
            walk(file, function (error) {
              next();
            });
          } else {
            var parser = mm(fs.createReadStream(file), function (err, metadata) {
              if (err) {
                next();
              }
              // Insert the object to the database
              db.collection("music").update({
                path: file
              }, {
                $set: {
                  "title": metadata.title,
                  "artist": metadata.artist,
                  "album": metadata.album
                }
              }, {
                upsert: true
              })
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

  console.log("Database updated.");
});


// Search, directed from aria.js AJAX
app.post('/search', function (req, res) {
  console.log("Received: " + JSON.stringify(req.body));

  // Database search
  MongoClient.connect(DB_URL, function (err, db) {
    if (err) {
      return console.log(err);
    }

    db.collection("music").find({
      $text: {
        $search: req.body.search
      }
    }).toArray(function (err, result) {
      if (err) throw err;
      console.log(result);
      res.send(result);
    });
    db.close();
  });
});

var server = app.listen(PORT, IP);