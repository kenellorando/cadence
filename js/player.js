function playerToggle() {
	var stream = document.getElementById("stream");
	
	if (stream.paused) {
		stream.play();
		document.getElementById("playerToggle").innerHTML = "❚❚";
	} else {
		location.reload();
		document.getElementById("playerToggle").innerHTML = "►";
	}
}

function volumeToggle(vol) {
	var volume = document.getElementById("stream");
	volume.volume = vol / 100;
}