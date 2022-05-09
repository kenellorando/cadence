
streamSrcURL = "" // this gets used by the stream playButton function

// Hook into the cadence radio data socket
$(document).ready(function() {
	var socket = new WebSocket("ws://" + location.host + "/api/aria1/radiodata/socket")

	socket.onopen = () => {
		console.log("Established connection with Cadence radiodata socket.")
	}
	socket.onmessage = (ServerMessage) => {
		handle(ServerMessage)
	}
	socket.onerror = (ServerMessage) => {
		console.warn("Could not reach the Cadence radio data socket: " + ServerMessage.data)
	}

	function handle(ServerMessage) {
		let message = JSON.parse(ServerMessage.data)
		switch (message.Type) {
			case "NowPlaying":
				var nowPlayingArtist = message.Artist.trim();
				var nowPlayingTitle = message.Title.trim();
				$('#artist').text(nowPlayingArtist);
				$('#song').text(nowPlayingTitle);
				console.log("Now playing: " + nowPlayingArtist + ", '" + nowPlayingTitle + "'")
				break;
			case "Listeners":
				var currentListeners =  message.Listeners;
				if (currentListeners == -1) {
					document.getElementById("listeners").innerHTML = "(stream unreachable)"
				} else {
					document.getElementById("listeners").innerHTML = currentListeners;
				}
				break;
			case "StreamConnection":
				var currentListenURL =  message.ListenURL.trim();
				var currentMountpoint = message.Mountpoint.trim();
				
				if (currentListenURL != "unknown") {
					$('#status').html("Connected to stream: <a href='"+ streamSrcURL + "'>" + currentMountpoint + "</a>");
				} else {
					$('#status').html("Disconnected from stream.");
				}
				document.getElementById("stream").src = currentListenURL
				streamSrcURL = currentListenURL // set global URL
				break;
		}
	}
});

// Toggle the stream with the playButton
$(document).ready(function() {
	var stream = document.getElementById("stream");
	var mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent);
	document.getElementById("playButton").addEventListener('click', function() {
		if (stream.paused) {
			stream.src = streamSrcURL;
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

// Get latest source release title
$(document).ready(function() {
	$.ajax({
		type: 'GET',
		url: "/api/aria1/version",
		dataType: "json",
		// On success, format data into table
		success: function(data) {
			document.getElementById("release").innerHTML = data.Version;
		},
		error: function() {
			document.getElementById("release").innerHTML = "(N/A)";
		}
	});
});
// Display page warning on iOS or Safari devices
$(document).ready(function() {
	let safariUA = /Apple/i.test(navigator.vendor);
	let iOSUA = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
	if (iOSUA || safariUA) {
		alert("You appear to be using an iOS device or a Safari browser. Cadence stream playback may not be compatible with your platform.")
	}
});
// Volume control
$(document).ready(function() {
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
