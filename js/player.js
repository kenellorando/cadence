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

var themeNames = [ // Place in array matches the themeID
	"chicago-evening",
	"cyberpunk-bartender"
	// Add additional themes here
];

// Called on page load to perform initialization and autoplay
function readyPlayer() {
	playerToggle(document.body);
	
	var volume;
	var theme;
	
	// Fetch the volume and themeID from the URL GET parameters passed earlier (or typed in a link, or by the user, or in a bookmark)
	// Interesting note, the fact that this can be bookmarked means that users can personalize setting defaults by bookmarking them.
	// First, the volume...
	var par = document.URL.match(/volume=([0-9]+)/);
	volume = par ? par[1] : 20; // volume defaults to 20%
	// Now, the theme.
	par = document.URL.match(/theme=([0-9]+)/);
	theme = par ? par[1] : 0; // themeID defaults to 0 (chicago-evening)
	
	// Set the volume
	var stream = document.getElementById("stream");
	var volumeControl = document.getElementById("volume");
	stream.volume = volume / 100;
	volumeControl.value = volume;
	
	// Set the theme by its ID.
	setThemeById(theme);
}

// Get the themeID of a theme by its name.
function themeID(themeName) {
	switch (themeName) {
		case "chicago-evening":
			return 0;
		case "cyberpunk-bartender":
			return 1;
		// Add additional themes here
		default: // Default back to chicago-evening
			return 0;
	}
}

// Set the page theme by a themeID
function setThemeById(themeID) {
	switch (themeID) {
		case 0:
			selectChicagoEvening();
			break;
		case 1:
			selectCyberpunkBartender();
			break;
		// Add additional themes here
		default: // Default back to chicago-evening
			selectChicagoEvening();
	}
}

// Convenience: Shorthand to set the page theme by theme name
function setThemeByName(themeName) {
	setThemeById(themeID(themeName));
}