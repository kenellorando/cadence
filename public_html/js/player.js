// I'll put all the default onload stuff in here
function defaultPlayer() {
  var vol = 0.77;
  document.getElementById("volume").value = 0.77;
  var volume = document.getElementById("stream");
  volume.volume = vol;
  
  // If on mobile, only preload stream metadata
  if (/Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent))
    document.getElementById("stream").preload="metadata";
}

// When you hit the play button
function playerToggle() {
  // Whether the browser claims to be a mobile device.
  var mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent);
  var stream = document.getElementById("stream");

  if (stream.paused) {
	if (mobile) {
	  // Set the stream source back to what it should be
	  stream.src = "http://198.37.25.127:8000/cadence1"; // Make sure this stays current, otherwise the stream will not resume playing
	}
	// Reload and play the stream
	stream.load();
    stream.play();
    document.getElementById("playerToggle").innerHTML = "❚❚";
  } else {
	if (mobile) {
	  // Set the stream source to an empty URL that claims to be an OGG
	  stream.src = URL.createObjectURL(new Blob([], {type:"application/ogg"}));
	  // And reload the stream (it is now paused, since it it playing nothing)
	  stream.load();
	}
	stream.load();
    document.getElementById("playerToggle").innerHTML = "►";
  }
}

// When you change the volume
function volumeToggle(vol) {
  var volume = document.getElementById("stream");
  volume.volume = vol;
}

// GETS and displays currently playing info
function radioTitle() {
  // Located on testament's stream web folder
  var url = 'http://198.37.25.127:8000/json.xsl';

  $.ajax({
    type: 'GET',
    url: url,
    async: true,
    jsonpCallback: 'parseMusic',
    contentType: "application/json",
    dataType: 'jsonp',
    success: function (json) {
      // do not mix up id with the "title" for the page heading
      $('#song_title').text(json['/cadence1']['song_title']);
      $('#artist_name').text(json['/cadence1']['artist_name']);
    },
    error: function (e) {
      console.log(e.message);
    }
  });
}

$(document).ready(function () {
  setTimeout(function () {
    radioTitle();
  }, 0);
  setInterval(function () {
    radioTitle();
  }, 10000);
});