// Switch the display of the front matter
$(document).ready(function () {
    document.getElementById("tabRequest").addEventListener('click', function(){
        replace("Request")
    }, true);

    document.getElementById("tabAbout").addEventListener('click', function(){
        replace("About")
    }, true);

    document.getElementById("tabTheme").addEventListener('click', function(){
        replace("Theme")
    }, true);

    document.getElementById("tabRequest").style.fontWeight = "600";
});

function replace(target) {
    keys = [ "About", "Request", "Theme" ]
    for (key of keys) {
        if (key == target) {
            document.getElementById("front" + key).style.display = "block";
            document.getElementById("tab" + key).style.fontWeight = "600";
        } else {
            document.getElementById("front" + key).style.display = "none";
            document.getElementById("tab" + key).style.fontWeight = "300";
        }
    }
}