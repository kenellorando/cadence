// Callbacks.js: This file contains all callback classes used in theme.json
// This file must be loaded before theme.json

// Call this function to register a callback class as inheriting from the CallbackInterface class
// This is required to ensure compliance with the interfaces used by the theme engine.
function registerCallback(child) {
    child.prototype=Object.create(CallbackInterface.prototype)
    child.prototype.constructor=child
}

// LoggingCallback: This is an example callback (to show syntax)
// This callback simply prints messages whenever callback functions are called.
function LoggingCallback(theme) {
    CallbackInterface.call(this, theme) // Call the parent constructor. This is highly recommended, but not strictly necessary.
}

registerCallback(LoggingCallback)

LoggingCallback.prototype.preLoad=function(currentTheme) {
    console.log("Ready to switch from "+(currentTheme!=null ? currentTheme.themeKey : "nothing")+" to "+this.theme.themeKey)
    return false
}

LoggingCallback.prototype.postLoad=function() {
    console.log("Switched to "+this.theme.themeKey)
}

LoggingCallback.prototype.preUnload=function(pendingTheme) {
    console.log("Ready to switch away from "+this.theme.themeKey+" to "+pendingTheme.themeKey)
}

LoggingCallback.prototype.postUnload=function () {
    console.log("Switched away from "+this.theme.themeKey)
}

LoggingCallback.prototype.nightmodeSwitch=function () {
    console.log("Switching theme into nightmode...")
}

LoggingCallback.prototype.daymodeSwitch=function () {
    console.log("Switching theme into daymode...")
}


// Cyberpunk's background switching callback
function CyberpunkCallback(theme) {
    CallbackInterface.call(this, theme)

    this.cancel=-1
    this.lastIndex=-1
}

registerCallback(CyberpunkCallback)

// Helper functions
CyberpunkCallback.prototype.cancelSwitcher=function () {
    clearTimeout(this.cancel);
    document.getElementsByTagName("html")[0].style.backgroundImage = "";
}

CyberpunkCallback.prototype.backgroundSwitcher=function () {
    // If we're in night-mode, do not perform any transition
    if (cyberpunkNight)
        return;

    var html = document.getElementsByTagName("html")[0];
    html.classList.remove("transition"); // Cancel any waiting transition

    // An array of all the filenames of possible backgrounds
    var URLs = [
    "https://i.imgur.com/SySfXrk.gif",
    "https://i.imgur.com/jwmBsvs.gif",
    "https://i.imgur.com/nyRGAM5.gif"]; // Nighttime background intentionally excluded.

    // Generate a random index, that isn't the last chosen one
    var index = this.lastIndex;
    do {
        index = Math.floor(Math.random() * URLs.length);
    } while (index == this.lastIndex);

    // And remember the last chosen one
    this.lastIndex = index;

    // The chosen URL
    var url = URLs[index]

    var maxTime = 120000; // Maximum time a background can stay: milliseconds
    var minTime = 10000; // Minimum time a background must stay: milliseconds
    var timeRange = maxTime-minTime;
    var time = Math.floor(Math.random() * timeRange) + minTime; // Random time within bounds

    // Cancel any waiting switch
    this.cancelSwitcher();

    // Switch the background
    html.style.backgroundImage = "url("+url+")";

    // And queue the fade animation to play after our chosen time
    this.cancel = setTimeout(this.fadeFromWhite, time);
}

CyberpunkCallback.prototype.fadeFromWhite=function() {
    var html = document.getElementsByTagName("html")[0];

    // Adding this class runs the animation
    html.classList.add("transition");

    // Choose a new background in 3 seconds, after the animation ends
    this.cancel = setTimeout(this.backgroundSwitcher, 3000);
}

// Callbacks
CyberpunkCallback.prototype.postLoad=function () {
    // On load, start the switching by fading in from white
    this.fadeFromWhite()
}

CyberpunkCallback.prototype.preUnload=function() {
    // Before unloading, cancel pending switches
    this.cancelSwitcher()
}

// Note that we don't implement all six callbacks. We inherit default behavior from CallbackInterface thanks to the magic of registerCallback
