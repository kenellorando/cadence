/*
 * Handlers for page events for Cadence
 * This file replaces inline handlers in the page with applications in jQuery
 */

// Set up the handlers when the DOM is ready for them
$(document).ready(function() {
    // Stylesheet elements which are marked as awaiting load
    var delayedCss=document.querySelectorAll(".delayedLoad");

    // Iterate over those sheets and set them to be applicable for loading
    for (var i=0; i<delayedCss.length; ++i) {
        if (delayedCss[i].media!="all") {
            delayedCss[i].media="all"
        }
    }
});
