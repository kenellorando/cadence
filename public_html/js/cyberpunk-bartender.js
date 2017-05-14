var cyberpunkCancel; // So that switches can be canceled on demand
var cyberpunkLast = -1; // Index of the last chosen background

var cyberpunkNight=0; // Are we in night mode?

// Cancels a waiting switch and clears a set image
function cancelSwitcher() {
    clearTimeout(cyberpunkCancel);
    document.getElementsByTagName("html")[0].style.backgroundImage = "";
}

// Runs the "fade from white" animation, then chooses a new background
function fadeFromWhite() {
    var html = document.getElementsByTagName("html")[0];

    // Adding this class runs the animation
    html.classList.add("transition");

    // Choose a new background in 3 seconds, after the animation ends
    cyberpunkCancel = setTimeout(cyberpunkBackgroundSwitcher, 3000);
}

// Handles the switching of the cyberpunk bartender background
function cyberpunkBackgroundSwitcher() {
    // If we're in night-mode, do not perform any transition
    if (cyberpunkNight)
        return;

    var html = document.getElementsByTagName("html")[0];
    html.classList.remove("transition"); // Cancel any waiting transition

    // An array of all the filenames of possible backgrounds
    var URLs = [
    "cyberpunk-bartender.gif",
    "cyberpunk-bartender2.gif",
    "cyberpunk-bartender3.gif"];

    // A URL to the folder in which backgrounds live
    // First, our current location....
    var URLroot = document.location.href;
    // Then, minus the filename (if one exists)
    URLroot = URLroot.substring(0, URLroot.lastIndexOf("/")+1);
    // And add the media folder
    URLroot += "media/";

    // Generate a random index, that isn't the last chosen one
    var index = cyberpunkLast;
    do {
        index = Math.floor(Math.random() * URLs.length);
    } while (index == cyberpunkLast);

    // And remember the last chosen one
    cyberpunkLast = index;

    // The chosen URL
    var url = URLroot+URLs[index]

    var maxTime = 120000; // Maximum time a background can stay: milliseconds
    var minTime = 10000; // Minimum time a background must stay: milliseconds
    var timeRange = maxTime-minTime;
    var time = Math.floor(Math.random() * timeRange) + minTime; // Random time within bounds

    // Cancel any waiting switch
    cancelSwitcher();

    // Switch the background
    html.style.backgroundImage = "url("+url+")";

    // And queue the fade animation to play after our chosen time
    cyberpunkCancel = setTimeout(fadeFromWhite, time);
}