// Helper function: Does the work of setting theme color for all meta tags
function setThemeColor(color) {
  document.getElementById("chrome-color").content = color;
  document.getElementById("ie-color").content = color;
}

// Handles clicks 
$(document).ready(function () {
  $('.themeChoice').on('click', function () {
    var themeChoice = $(this).attr('id');
    themeChanger(themeChoice);
  });
});

function themeChanger(themeName) {
  // Returns the specific theme as a single object
  var themeObj = theme[themeName];
  console.log(themeObj);

  document.getElementById("selected-css").href = themeObj.cssPath;
  document.getElementById("title").innerHTML = themeObj.title;
  document.getElementById("subtitle").innerHTML = themeObj.subtitle;
  setThemeColor(themeObj.themeColor);
  localStorage.setItem('themeKey', themeObj.themeKey);


  // CYBERPUNK *************************
  if (themeName === "cyberpunkBartender") {
    var currentHour = new Date().getHours();
    // IF condition states the daytime hours
    // 8:00:00 PM - 9:59:59 AM
    if (currentHour >= 8 && currentHour < 22) {
      changeTo = {
        css: "/css/themes/cyberpunk-bartender.css",
        title: "CADENCE",
        subtitle: "A Retro Cyberpunk Jukebox",
        themeColor: "#B30E67", // Hot pink
        themeKey: "cyberpunkBartender"
      };
    } else {
      changeTo = {
        css: "/css/themes/cyberpunk-bartender-night.css",
        title: "CADENCE",
        subtitle: "A Retro Cyberpunk Jukebox",
        themeColor: "#1D2951", // Navy
        themeKey: "cyberpunkBartender"
      };
    }
  }

  // LIGHT MAGE ***********************
  else if (themeName === "lightMage") {
    changeTo = {
      css: "/css/themes/light-mage.css",
      title: "CADENCE",
      subtitle: "Just An Ordinary Radio",
      themeColor: "#FFFFFF", // White
      themeKey: "lightMage"
    };

    var video = document.getElementById("video-source");
    // Quick and dirty fix to get absolute URL to fix the stuttering background
    var filename = document.location + "/media/lux1.mp4";
    // Loads the video source
    if (video.src !== filename) {
      video.src = filename;
      video.parentElement.load(); // The parent element of video is the div "fullscreen-bg"
    }
  }
}

// This is run onload. To change the default theme, (for users that have not yet picked one) change the statement for null
function defaultTheme() {
  var theme = localStorage.getItem('themeKey');
  if (theme === null) {
    themeChanger("chicagoEvening");
  } else {
    themeChanger(theme);
  }
}

// Reselects for time-based themes at a set interval
window.setInterval(function () {
  defaultTheme();
}, 10000);
