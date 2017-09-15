const PORT = 8080;
const IP = '198.37.25.185';

var express = require('express'),
app = express(),
path = require('path');




app.get('/', function(req, res) {
  console.log("Data:" + req.query['song']);
})








app.use(express.static(path.join(__dirname, 'public')));

var server = app.listen(PORT, IP);
