// Hook into the cadence nowPlaying socket
$(document).ready(function() {
	var socket = new WebSocket("ws://" + location.host + "/api/aria1/nowplaying/socket")

	socket.onopen = () => {
		console.log("Connected to Cadence nowPlaying socket.")
	}
	socket.onmessage = (ServerMessage) => {
		console.log("Now playing: " + ServerMessage)
		updateNowPlaying(ServerMessage)
	}

	function updateNowPlaying(ServerMessage) {
		let song = JSON.parse(ServerMessage.data)
			var nowPlayingArtist = song['Artist'].trim();
			var nowPlayingTitle = song['Title'].trim();
			$('#artist').text(nowPlayingArtist);
			$('#song').text(nowPlayingTitle);
	}
});

streamSrcURL = "" // this gets used by the stream playButton function
// Hook into the cadence streamURL socket
$(document).ready(function() {
	var socket = new WebSocket("ws://" + location.host + "/api/aria1/streamurl/socket")

	socket.onopen = () => {
		console.log("Connected to Cadence streamURL socket.")
	}
	socket.onmessage = (ServerMessage) => {
		console.log("New stream URL: " + ServerMessage.data)
		updateStreamURL(ServerMessage)
	}

	function updateStreamURL(ServerMessage) {
		let stat = JSON.parse(ServerMessage.data)
		var currentListenURL = stat['ListenURL'].trim();
		var currentMountpoint = stat['Mountpoint'].trim();

		var stream = document.getElementById("stream");
		stream.src = currentListenURL;
		streamSrcURL = currentListenURL // set global URL
	
		if (currentListenURL != "unknown") {
			$('#status').html("Connected to stream: <a href='"+ streamSrcURL + "'>" + currentMountpoint + "</a>");
		} else {
			$('#status').html("Disconnected from stream.");
		}
	}
});

// Hook into the cadence streamListeners socket
$(document).ready(function() {
	var socket = new WebSocket("ws://" + location.host + "/api/aria1/streamlisteners/socket")

	socket.onopen = () => {
		console.log("Connected to Cadence streamListeners socket.")
	}
	socket.onmessage = (ServerMessage) => {
		console.log("Listener count update: " + ServerMessage.data)
		updateStreamListeners(ServerMessage)
	}

	function updateStreamListeners(ServerMessage) {
		let stat = JSON.parse(ServerMessage.data)
		var currentListeners = stat['Listeners'];
		var listeners = document.getElementById("listeners");
		if (currentListeners == -1) {
			listeners.innerHTML = "(stream unreachable)"
		} else {
			listeners.innerHTML = currentListeners;
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
		url: 'https://api.github.com/repos/kenellorando/cadence/releases/latest',
		// On success, format data into table
		success: function(data) {
			document.getElementById("release").innerHTML = data.name;
		},
		error: function() {
			document.getElementById("release").innerHTML = "Could not retrieve version data.";
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