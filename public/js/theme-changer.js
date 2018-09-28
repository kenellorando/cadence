// Helper function: Does the work of setting theme color for all meta tags
function setThemeColor(color) {
  document.getElementById("chrome-color").content = color;
  document.getElementById("ie-color").content = color;
}

// Function that starts playing the video for video themes
function setVideo(themeObj) {
  var video = document.getElementById("video-source");
  var filename = document.location.origin + themeObj.videoPath;
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

var callback = new CallbackInterface(null);

// The main theme setter function
function themeChanger(themeName) {
  // Returns the specific theme as a single object
  var themeObj = Object.assign({}, theme[themeName]);
  var currentHour = new Date().getHours();

  var t=themeObj;

  // Copy of callback to be called in posthooks
  var call=Object.assign({}, callback);
  Object.setPrototypeOf(call, callback.constructor.prototype);

  do {
      // If the theme is blocked on mobile, and we're on mobile, default to chicagoEvening
      // Uses the same mobile check as Ken and I used back in the beginning, which is still used for pause
      if (t.blockMobile && /Android|webOS|iPhone|iPad|iPod|BlackBerry/i.test(navigator.userAgent)) {
          var name=t.mobileTheme || 'chicagoEvening';
          themeObj=theme[name];
      }
      else {
          themeObj=t;
      }

      // Setup the theme's callback
      themeObj.callback = themeObj.callback || CallbackInterface;
      themeObj.callback = new themeObj.callback.prototype.constructor(themeObj);
  } while (t=themeObj.callback.preLoad(callback.theme));

  callback.preUnload(themeObj);

  // If a nightmode exists
  if (themeObj.hasNightMode == true) {
    // If it is nighttime
    if (currentHour < 8 || currentHour > 22) {
      themeObj.callback.nightmodeSwitch();
      themeNameNight = themeName + "Night";
      var themeObjNight;
      t = Object.assign({}, theme[themeNameNight]);
      themeObj.callback.preUnload(t);
      themeObj.callback.postUnload();
      do {
          themeObjNight = t;
          themeObjNight.callback = themeObjNight.callback || CallbackInterface;
          themeObjNight.callback = new themeObjNight.callback.prototype.constructor(themeObjNight);
      } while (t=themeObjNight.callback.preLoad(callback.theme));
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
      setTimeout(function() {
          themeObjNight.callback.daymodeSwitch();
          defaultTheme();
      }, diff); // Schedule a theme default for one second after 9 AM

      // Call the post hooks as soon as the thread becomes idle
      setTimeout(function() {
          call.postUnload();
          themeObjNight.callback.postLoad();
      }, 0)

      callback=themeObjNight.callback;
      return
    }
    // Else, set daytime and schedule a theme reset shortly after nighttime
    else {
      document.getElementById("selected-css").href = themeObj.cssPath;
      document.getElementById("title").innerHTML = themeObj.title;
      document.getElementById("subtitle").innerHTML = themeObj.subtitle;
      setThemeColor(themeObj.themeColor);
      localStorage.setItem('themeKey', themeObj.themeKey);
      if (themeObj.videoPath) {
        setVideo(themeObj);
      }

      var time=new Date();
      var target=new Date();
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

  // Call the post hooks as soon as the thread becomes idle
  setTimeout(function() {
      call.postUnload();
      themeObj.callback.postLoad();
  }, 0);

  callback=themeObj.callback;
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

// Execute the load handler whenever the page is ready
$(document).ready(defaultTheme);
