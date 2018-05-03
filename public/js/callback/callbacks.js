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
