$(document).ready(function() {
	// Load audio stream source
	var stream = document.getElementById("stream");
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

	// Warn iOS and Safari users
	let safariUA = /Apple/i.test(navigator.vendor);
	let iOSUA = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
	if (iOSUA || safariUA) {
		alert("You appear to be using an iOS device or a Safari browser. Cadence stream playback may not be compatible with your platform.")
	}

	// Volume control
	// Load cached volume level, or 30%
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

	// Tab switching
	$('#tabs li').on('click', function() {
		var tab = $(this).data('tab');
	
		$('#tabs li').removeClass('is-active');
		$(this).addClass('is-active');
	
		$('#tab-content section').removeClass('is-active');
		$('section[data-content="' + tab + '"]').addClass('is-active');
	});
	
	// Search keyup
	$("#searchInput").keyup(function(event) {
		if (event.keyCode == 13) {
			postSearch()
		}
	});
});
