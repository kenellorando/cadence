"use strict";
// Helper function: Does the work of setting theme color for all meta tags
function setThemeColor(color) {
  document.getElementById("chrome-color").content = color;
  document.getElementById("ie-color").content = color;
}

$(document).ready(function () {
  $('.themeChoice').on('click', function () {
    var themeChoice = $(this).attr('id');
    themeChanger(themeChoice);
  });
});


function themeChanger(themeName) {
  var changeTo = {};

  // CHICAGO EVENING ******************
  if (themeName === "chicagoEvening") {
    changeTo = {
      css: "/css/themes/chicago-evening.css",
      title: "CADENCE",
      subtitle: "A Rhythmic Experience",
      themeColor: "#1d2951", // Navy
      themeKey: "chicagoEvening"
    };
  }

  // CYBERPUNK *************************
  else if (themeName === "cyberpunkBartender") {
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

  // MAYBERRY ***************************
  else if (themeName === "mayberry") {
    changeTo = {
      css: "/css/themes/mayberry.css",
      title: "CADENCE",
      subtitle: "A Rhythmic Ξxperience",
      themeColor: "#000000", // Black
      themeKey: "mayberry"
    };
  }

  // ELECTROMASTER **********************
  else if (themeName === "electromaster") {
    changeTo = {
      css: "/css/themes/electromaster.css",
      title: "ケイデンス",
      subtitle: "A Certain Scientific Radio",
      themeColor: "#09C1FF", // A certain scientific light blue
      themeKey: "electromaster"
    };
  }

  // STAR GUARDIAN *******************
  else if (themeName === "starGuardian") {
    changeTo = {
      css: "/css/themes/star-guardian.css",
      title: "<span id='ke'>ケ</span><span id='i'>イ</span><span id='de'>デ</span><span id ='n'>ン</span><span id='su'>ス</span>",
      subtitle: "A Stellar Experience",
      themeColor: "#ecb1e9", // Lux pink
      themeKey: "starGuardian"
    };
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

  // YORHA ******************************
  else if (themeName === "yorha") {
    changeTo = {
      css: "/css/themes/yorha.css",
      title: "CADENCE",
      subtitle: "For The Glory of Mankind",
      themeColor: "#000000", // Black
      themeKey: "yorha"
    };
  }

  // Change HTML to changeTo's properties
  document.getElementById("selected-css").href = changeTo.css;
  document.getElementById("title").innerHTML = changeTo.title;
  document.getElementById("subtitle").innerHTML = changeTo.subtitle;
  setThemeColor(changeTo.themeColor); // A dark navy
  localStorage.setItem('themeKey', changeTo.themeKey);
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
}, 1000);
