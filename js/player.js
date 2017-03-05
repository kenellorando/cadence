function playerToggle(e) {
  var stream = document.getElementById("stream");

  if (stream.paused) {
    stream.play();
    document.getElementById("playerToggle").innerHTML = "❚❚";
  } else {
    stream.pause();
	e.onclick = function () { // The mindblowing workaround to a desyncing pause button
		var loc = document.location.pathname; // Get the path to the document, without parameters
		loc += "?volume="+document.getElementById("volume").value; // Append the currently set volume
		document.location = loc; // Reload the page with that volume set.
	};
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
	
	var volume;
	
	// Fetch the volume from the URL GET parameters
	var par = document.URL.match(/volume=([0-9]+)/);
	volume = par ? par[1] : 20; // volume defaults to 20%
	
	// Set the volume
	var stream = document.getElementById("stream");
	var volumeControl = document.getElementById("volume");
	stream.volume = volume / 100;
	volumeControl.value = volume;
}