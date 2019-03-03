// Displays currently playing info, from the Icecast xsl
function radioTitle() {
    var url = 'http://stream.cadenceradio.com:8000/now-playing.xsl';
    $.ajax({
        type: 'GET',
        url: url,
        async: true,
        jsonpCallback: 'parseMusic',
        contentType: "application/json",
        dataType: 'jsonp',
        success: function (json) {
            // Grab and trim song data
            var serverName = json['/cadence1']['server_name'].trim();
            var nowPlayingArtist = json['/cadence1']['artist_name'].trim();
            var nowPlayingSong = json['/cadence1']['song_title'].trim();
            
            // Set info in the player
            $('#status').text("Connected to server: " + serverName)
            $('#song').text(nowPlayingSong);
            $('#artist').text(nowPlayingArtist);
            
            // Set the browser title to the now playing info
            window.document.title =  "CR♥ | " + nowPlayingArtist + " - " + nowPlayingSong;
        },
        error: function (e) {
            console.log(e.message);
            $('#status').text("Disconnected from server.")
            $('#song').text("-");
            $('#artist').text("-");
            document.getElementById("status").innerHTML = "Disconnected from server."
            document.getElementById("artist").innerHTML = "-";
            document.getElementById("song").innerHTML = "-";
        }
    })
};

// Toggle the stream with the playButton
$(document).ready(function () {
    document.getElementById("playButton").addEventListener('click', function(){
        var stream = document.getElementById("stream");
        var mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent);

        // Here, we pause and play the stream
        // If the device is mobile, we remove the stream source entirely
        // so music data stops loading in the background.
        if (stream.paused) {
            // Reload the audio source if on mobile
            if (mobile) {
                stream.src = "http://stream.cadenceradio.com:8000/cadence1";
            }
            stream.load();
            stream.play();
            // Replace the ❙❙ in the button when playing
            document.getElementById("playButton").innerHTML = "❙❙";
        } else {
            // If mobile, clear the audio source
            if (mobile) {
                stream.src = "";
            }
            stream.load();
            // Replace the ► in the button when paused
            document.getElementById("playButton").innerHTML = "►";
        }
    }, true);
});

// Update now playing info at an interval
$(document).ready(function () {
    setTimeout(function () {
        radioTitle();
    }, 0);
    setInterval(function () {
        radioTitle();
    }, 10000);
});