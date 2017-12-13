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

  // If a nightmode exists
  if (themeObj.hasNightMode == true) {
    // If it is nighttime
    if (currentHour < 8 || currentHour > 22) {
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
      
      // Schedule a theme reset shortly after daytime
      var time=new Date();
      var target=new Date();
      Object.assign(target,time);
      target.setHours(9);
      target.setMinutes(0);
      target.setSeconds(1); // Offset by one second just in case
      // Offset the day if its after 8 [since we can't be here if its not nighttime]
      if (time.getHours()>20) {
          target.setDate(target.getDate()+1);
      }
      var diff=target-time; // Milliseconds
      setTimeout(defaultTheme, diff); // Schedule a theme default for one second after 9 AM
    }
    // Else, schedule a theme reset shortly after nighttime
    else {
      var time=new Date();
      var target=new Object();
      Object.assign(target,time);
      target.setHours(23);
      target.setMinutes(0);
      target.setSeconds(1); // Offset by one second just in case
      var diff=target-time; // Milliseconds
      setTimeout(defaultTheme, diff); // Schedule a theme default for one second after 11 PM
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


// This is run onload. To change the default theme, (for users that have not yet picked one) change the statement for null
function defaultTheme() {
  var theme = localStorage.getItem('themeKey');
  if (theme === null) {
    themeChanger("chicagoEvening");
  } else {
    themeChanger(theme);
  }
}
