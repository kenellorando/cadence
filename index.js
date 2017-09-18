const PORT = 8080;
const IP = '198.37.25.185';

var express = require('express');
var path = require('path');
var bodyParser = require('body-parser');
var app = express();


var urlencodedParser = bodyParser.urlencoded({
    extended: false
});


app.use(express.static(path.join(__dirname, 'public')));


app.post('/search', urlencodedParser, function (req, res) {
    console.log(req.body);
});




var server = app.listen(PORT, IP);