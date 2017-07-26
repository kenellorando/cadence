function selectChicagoEvening() {
  cancelSwitcher();

  document.getElementById("selected-css").href = "/css/themes/chicago-evening.css";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Experience";
  localStorage.setItem('themeKey', 'chicagoEvening');
}

// Handles clicks
$(document).ready(function () {
  $('.themeChoice').on('click', function () {
    var themeChoice = $(this).attr('id');
    themeChanger(themeChoice);
  });
});

function themeChanger(themeName) {
  var currentHour = new Date().getHours();

  // If a nightmode exists and it is nighttime
  if (themeObj.hasNightMode == true && (currentHour < 8 || currentHour > 22)) {
    cyberpunkNight = true; // Inform the cyberpunk switcher that its nighttime
    themeNameNight = themeName + "Night";
    var themeObjNight = theme[themeNameNight];
    document.getElementById("selected-css").href = themeObjNight.cssPath;
    document.getElementById("title").innerHTML = themeObjNight.title;
    document.getElementById("subtitle").innerHTML = themeObjNight.subtitle;
    setThemeColor(themeObjNight.themeColor);
    localStorage.setItem('themeKey', themeObjNight.themeKey);
    // If the nightmode is a video theme
    if (themeObjNight.videoPath) {
      setVideo(themeObjNight);
    }
  }
  // Otherwise, no nightmode to fall back on
  else {
    cyberpunkNight = false; // Inform the cyberpunk switcher that its daytime. Or, that the current theme doesn't support nightmode, which means it doesn't matter.
    document.getElementById("selected-css").href = themeObj.cssPath;
    document.getElementById("title").innerHTML = themeObj.title;
    document.getElementById("subtitle").innerHTML = themeObj.subtitle;
    setThemeColor(themeObj.themeColor);
    localStorage.setItem('themeKey', themeObj.themeKey);
    if (themeObj.videoPath) {
      setVideo(themeObj);
    }
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