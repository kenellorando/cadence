const PORT = 8080;
const IP = 'localhost';

var express = require('express');
var path = require('path');
var bodyParser = require('body-parser');
var app = express();


// Parse incoming data
app.use(bodyParser.urlencoded({
  extended: true
}));

// Point to publicly served files
app.use(express.static(path.join(__dirname, 'public')));


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
