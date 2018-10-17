// Callbacks.js: This file contains all callback classes used in theme.json
// This file must be loaded before theme.json

// Call this function to register a callback class as inheriting from the CallbackInterface class
// This is required to ensure compliance with the interfaces used by the theme engine.
function registerCallback(child) {
    child.prototype=Object.create(CallbackInterface.prototype)
    child.prototype.constructor=child
}

// Call this function to register a callback class as inheriting from a different class
// This class should either be registered through this function or through registerCallback
// Unless, of course, the class is CallbackInterface, in which case this is equivalent to registerCallback
// Note that calling this does not make the child call the parent's callback functions.
// You should take care of this yourself.
function registerCallbackAs(child, parent) {
    child.prototype=Object.create(parent.prototype)
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

// CodecCallback: Automatically generates extra source tags for HEVC (.mkv) and VP9 (.webm) sources
// For video themes which benefit from these sources.
// Will prevent load if the theme does not have a videoPath.
function CodecCallback(theme) {
    CallbackInterface.call(this, theme);
}

registerCallback(CodecCallback);

// Insert the new source tags into the page
CodecCallback.prototype.preLoad=function () {
    // Check for null videoPath
    // (Which breaks this callback)
    if (this.theme.videoPath===undefined) {
        // This won't do.
        // Load some other theme - We don't care which.
        return true;
    }

    // Filename except extension
    var prefix=this.theme.videoPath.substring(0, this.theme.videoPath.lastIndexOf('.'));

    // We want to insert two tags, both before the existing video source
    // First, we prefer the HEVC media if the browser understands it.
    var avc=document.getElementById('video-source');
    var video=avc.parentElement;

    var hevc=document.createElement("source");
    hevc.id="video-hevc";
    hevc.src=prefix+".mkv";
    hevc.type="video/x-matroska; codecs=hevc";

    // If the browser doesn't like HEVC, use VP9 (webm).
    var webm=document.createElement("source");
    webm.id="video-webm";
    webm.src=prefix+".webm";
    webm.type="video/webm; codecs=vp9";

    // Append both children to the video elementFromPoint
    video.insertBefore(webm, avc);
    video.insertBefore(hevc, webm);

    // The browser may now load the theme.
    return false;
}

// Remove the source tags from the page before we switch to another theme
CodecCallback.prototype.preUnload=function(pendingTheme) {
    document.getElementById("video-hevc").remove();
    document.getElementById("video-webm").remove();
}
