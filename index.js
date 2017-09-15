const PORT = 8080;
const IP = '198.37.25.185';

var express = require('express'),
app = express(),
path = require('path');




app.post('./test', function(req, res) {
  var input = req.body.song;
  console.log("Received" + input);
});





app.use(express.static(path.join(__dirname, 'public')));

var server = app.listen(PORT, IP);
