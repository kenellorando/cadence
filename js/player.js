function playerToggle(e) {
  var stream = document.getElementById("stream");

  if (stream.paused) {
    stream.play();
    document.getElementById("playerToggle").innerHTML = "❚❚";
  } else {
    stream.pause();
	e.onclick = function () {
		document.location.reload();
	}; // The mindblowing workaround to a desyncing pause button
    document.getElementById("playerToggle").innerHTML = "►";
  }
}

function volumeToggle(vol) {
  var volume = document.getElementById("stream");
  volume.volume = vol / 100;
}

// Called on page load to perform initialization and autoplay
function readyPlayer() {
	playerToggle(document.body);
	
	// Set the volume
	var stream = document.getElementById("stream");
	var volumeControl = document.getElementById("volume");
	stream.volume = 0.2; // For now, default to 20% volume
	volumeControl.value = 20;
}