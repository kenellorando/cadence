// Helper function: Does the work of setting theme color for all meta tags
function setThemeColor(color) {
  document.getElementById("chrome-color").content = color;
  document.getElementById("ie-color").content = color;
}

// Function that starts playing the video for video themes
function setVideo(themeObj) {
  var video = document.getElementById("video-source");
  var filename = document.location + themeObj.videoPath;
  // Loads the video source
  if (video.src !== filename) {
    video.src = filename;
    video.parentElement.load(); // The parent element of video is the div "fullscreen-bg"
  }
}

// Handles clicks 
$(document).ready(function () {
  $('.themeChoice').on('click', function () {
    var themeChoice = $(this).attr('id');
    themeChanger(themeChoice);
  });
});

// The main theme setter function
function themeChanger(themeName) {
  // Returns the specific theme as a single object
  var themeObj = theme[themeName];
  var currentHour = new Date().getHours();

  // If a nightmode exists and it is nighttime
  if (themeObj.hasNightMode == true && (currentHour < 8 || currentHour > 22)) {
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
    document.getElementById("selected-css").href = themeObj.cssPath;
    document.getElementById("title").innerHTML = themeObj.title;
    document.getElementById("subtitle").innerHTML = themeObj.subtitle;
    setThemeColor(themeObj.themeColor);
    localStorage.setItem('themeKey', themeObj.themeKey);
    if (themeObj.videoPath) {
      setVideo(themeObj);
    }
  }
}

// To change the default theme, (for users that have not yet picked one) change the statement for null
$(document).ready(function() {
  var theme = localStorage.getItem('themeKey');
  if (theme === null) {
    themeChanger("chicagoEvening");
  } else {
    themeChanger(theme);
  }
});

// Reselects for time-based themes at a set interval
window.setInterval(function () {
  defaultTheme();
}, 10000);