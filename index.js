const PORT = 8080;
const IP = 'localhost';

var express = require('express');
var path = require('path');
var bodyParser = require('body-parser');
var app = express();

app.use(bodyParser.urlencoded({extended: true}));

// Point to publicly served files
app.use(express.static(path.join(__dirname, 'public')));


app.post('/search', function (req, res) {
    // Only visible on the web server console
    var reqBody = JSON.stringify(req.body);
    console.log("Received: " + reqBody);

    // Pass response back
    res.send("OK RESPONSE FROM ARIA BACKEND");
    res.end();
});



var server = app.listen(PORT, IP);