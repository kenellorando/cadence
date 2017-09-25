const PORT = 8080;
const IP = '198.37.25.185';
const DB_PORT = 27017;
const DB_IP = 'localhost';

var express = require('express');
var path = require('path');
var bodyParser = require('body-parser');
var MongoClient = require('mongodb').MongoClient;

var app = express();


// Parse incoming data
app.use(bodyParser.urlencoded({
  extended: true
}));

// Point to publicly served files
app.use(express.static(path.join(__dirname, 'public')));

// Database connect
MongoClient.connect('http://' + DB_IP + ":" + DB_PORT, (err, database) => {
  if (err) {
    return console.log(err);
  }

  console.log("Connection established to database.");
});


// Search, directed from aria.js AJAX
app.post('/search', function (req, res) {
  console.log("Received: " + JSON.stringify(req.body));
  // Web Server Console:
  // Received: {"search":"railgun"}

  // TODO: Use req data here

  res.send("OK from ARIA!");
  res.end();
});

var server = app.listen(PORT, IP);