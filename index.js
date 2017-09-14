var express = require('express'),
app = express(),
path = require('path');

app.use(express.static(path.join(__dirname, 'public')));

var server = app.listen(8080, '198.37.25.185');
