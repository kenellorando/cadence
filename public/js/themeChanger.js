// themeButton click handling
$(document).ready(function() {
	$('.themeButton').on('click', function() {
		var themeChoice = $(this).attr('id');
		themeChanger(themeChoice);
	});
});
// Change theme
function themeChanger(themeChoice) {
	// theme object is from themeMetadata.js
	var targetTheme = theme[themeChoice];
	document.getElementById("themeStylesheet").href = targetTheme.css;
	if (targetTheme.videoSource) {
		setVideo(targetTheme)
	} else {
		document.getElementById("videoSource").src = "."
	}
	localStorage.setItem('themeKey', targetTheme.key);
	colorButton(themeChoice);
}
function setVideo(themeObj) {
	var video = document.getElementById("videoSource");
	var filename = document.location + themeObj.videoSource;
	// Loads the video source
	if (video.src !== filename) {
	  video.src = filename;
	  video.parentElement.load(); // The parent element of video is the div "fullscreen-bg"
	}
}
// Theme button functionality
function colorButton(themeChoice) {
	// Change all theme buttons colors to the inactive style
	var buttons = document.getElementsByClassName("themeButton");
	for (var i = 0, il = buttons.length; i < il; i++) {
		buttons[i].classList.remove("activeTheme");
		buttons[i].classList.add("inactiveTheme");
	}
	// Override the selected theme button the active style
	document.getElementById(themeChoice).classList.remove("inactiveTheme");
	document.getElementById(themeChoice).classList.add("activeTheme");
}
// Retrieve themeKey in localStoprage
$(document).ready(function() {
	var themeKey = localStorage.getItem('themeKey');
	if (themeKey === null) {
		themeChanger("Default");
	} else {
		themeChanger(themeKey);
	}
});