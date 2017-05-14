function selectChicagoEvening() {
  cancelSwitcher();

  document.getElementById("selected-css").href = "/css/themes/chicago-evening.css";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Experience";
  localStorage.setItem('themeKey', 'chicagoEvening');
}

function selectCyberpunkBartender() {
  cancelSwitcher();

  document.getElementById("title").innerHTML = "CADEN<span>C</span>E";
  document.getElementById("subtitle").innerHTML = "A Retro Cyberpunk Jukebox";

  var currentHour = new Date().getHours();

  // IF condition states the daytime hours
  // 8:00:00 PM - 9:59:59 AM
  if (currentHour >= 8 && currentHour < 22) {
    document.getElementById("selected-css").href = "/css/themes/cyberpunk-bartender.css";
    cyberpunkNight=true;
  } else {
    document.getElementById("selected-css").href = "/css/themes/cyberpunk-bartender-night.css";
    cyberpunkNight=false;
  }

  localStorage.setItem('themeKey', 'cyberpunkBartender');

  cyberpunkBackgroundSwitcher();
}

function selectSpaceStation() {
  cancelSwitcher();
 
  document.getElementById("selected-css").href = "/css/themes/space-station.css";
  document.getElementById("subtitle").innerHTML = "A Space Odyssey";
  localStorage.setItem('themeKey', 'iss');

  // Not sure how to get this to work. Want to keep the src empty until this is activated. Then remove it when another is selected
  var video = document.getElementById("video-source");

  // Loads the video source
  if (video.src != "http://kenellorando.com/media/iss.mp4") {
    video.src = "http://kenellorando.com/media/iss.mp4";
    video.parentElement.load(); // The parent element of video is the div "fullscreen-bg"
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