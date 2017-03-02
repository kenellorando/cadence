function playerToggle() {
	var stream = document.getElementById("stream");

	if (stream.paused) {
		stream.play();
		document.getElementById("playerToggle").innerHTML = "❚❚";
	} else {
		location.reload(); // The mindblowing workaround to a desyncing pause button
		document.getElementById("playerToggle").innerHTML = "►";
	}
}

function volumeToggle(vol) {
	var volume = document.getElementById("stream");
	volume.volume = vol / 100;
}