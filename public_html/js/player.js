// I'll put all the default onload stuff in here
function defaultPlayer() {
  // Selects either the localstorage volume or a default value
  var vol = localStorage.getItem('volumeKey') || 0.77;
  document.getElementById("volume").value = vol;
  var volume = document.getElementById("stream");
  volume.volume = vol;
}

// When you hit the play button
function playerToggle() {
  var stream = document.getElementById("stream");
  var mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent);
  
  if (stream.paused) {
    // Loads up the real stream again if mobile
    if (mobile) {
      stream.src = "http://198.37.25.127:8000/cadence1";
    }
    stream.load();
    stream.play();
    document.getElementById("playerToggle").innerHTML = "❚❚";
  } else {
    // Loads up nothing if mobile
    if (mobile) {
      stream.src = "";
    }
    stream.load();
    document.getElementById("playerToggle").innerHTML = "►";
  }
}

// When you change the volume
function volumeToggle(vol) {
  var volume = document.getElementById("stream");
  volume.volume = vol;
  
  // Sets the new set volume into localstorage
  localStorage.setItem('volumeKey', vol);
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