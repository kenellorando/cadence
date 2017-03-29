function selectChicagoEvening() {
  document.getElementById("selected-css").href = "/css/themes/chicago-evening.css";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Experience";
  localStorage.setItem('themeKey', 'chicagoEvening');
}

function selectCyberpunkBartender() {
  document.getElementById("selected-css").href = "/css/themes/cyberpunk-bartender.css";
  document.getElementById("subtitle").innerHTML = "A Retro Cyberpunk Jukebox";
  localStorage.setItem('themeKey', 'cyberpunkBartender');
}

function selectSpaceStation() {
  document.getElementById("selected-css").href = "/css/themes/space-station.css";
  document.getElementById("subtitle").innerHTML = "A Space Odyssey";
  localStorage.setItem('themeKey', 'iss');
 
  // Not sure how to get this to work. Want to keep the src empty until this is activated. Then remove it when another is selected
  var video = document.getElementById("video-source");
  var current = document.location.href; // The URL of the current document
  current = current.substring(0, current.lastIndexOf("/")+1); // The URL of the current document's folder
  var source = new URL("media/iss.mp4", current);
  if (video.src != source) { // Force-load the video iff it is not already being played.
	  video.src = source;
	  video.parentElement.load();
  }
}

// This is run onload. To change the default theme, (for users that have not yet picked one) change the statement for null
function defaultTheme() {
  var theme = localStorage.getItem('themeKey');
  if (theme === "chicagoEvening") {
    selectChicagoEvening();
  } else if (theme === "cyberpunkBartender") {
    selectCyberpunkBartender();
  } else if (theme === "iss") {
    selectSpaceStation();
  } else if (theme === null) {
    selectSpaceStation();
  }
}