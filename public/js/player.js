// Displays currently playing info, from the Icecast xsl
function radioTitle() {
    var url = 'https://stream.cadenceradio.com/now-playing.xsl';
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
            var listeners = json['/cadence1']['listeners'].trim();
             
            // Set info in the player
            $('#status').html("Connected to server: <a href='https://melody.systems' target='_blank'>" + serverName + "</a>");
            $('#song').text(nowPlayingSong);
            $('#artist').text(nowPlayingArtist);
            $('#listeners').text(listeners)
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
    var stream = document.getElementById("stream");
    var mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent);
    
    document.getElementById("playButton").addEventListener('click', function(){
        if (stream.paused) {
            stream.src = "https://stream.cadenceradio.com/cadence1";
            stream.load();
            stream.play();
            // Replace the ❙❙ in the button when playing
            document.getElementById("playButton").innerHTML = "❙❙";
        } else {
            // Clear the audio source
            stream.src = "";
            stream.load();
            stream.pause();
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


// Get latest source release title
$(document).ready(function () {
    $.ajax({
        type: 'GET',
        url: 'https://api.github.com/repos/kenellorando/cadence/releases/latest',
        // On success, format data into table
        success: function (data) {
            document.getElementById("release").innerHTML = data.name;
        },
        error: function () {
            document.getElementById("release").innerHTML = "Could not retrieve version data.";
        }
    });
});

// Display page warning on iOS or Safari devices
$(document).ready(function () {
    let safariUA = /Apple/i.test(navigator.vendor);
    let iOSUA = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;

    if (iOSUA || safariUA) {
        alert("You appear to be using an iOS device or a Safari browser. Cadence stream playback may not be compatible with your platform.")
    }
});

// Volume control
$(document).ready(function () {
    // Load cached volume level, or 30%
    // Frontend default maximum volume is 60% max source volume
    var vol = localStorage.getItem('volumeKey') || 0.30;
    document.getElementById("volume").value = vol;
    // Set active volume on audio stream to loaded value
    var volume = document.getElementById("stream");
    volume.volume = vol;

    // Volume bar listeners
    $("#volume").change(function() {
        volumeToggle(this.value);
    }).on("input", function() {
        volumeToggle(this.value);
    });

    // Volume control
    function volumeToggle(vol) {
        var volume = document.getElementById("stream");
        volume.volume = vol;
    
        // Sets the new set volume into localstorage
        localStorage.setItem('volumeKey', vol);
    }
});
