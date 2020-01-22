// Switch the display of the front matter
$(document).ready(function () {
    document.getElementById("tabMain").addEventListener('click', function(){
        document.getElementById("frontMain").style.display = "block";
        document.getElementById("frontRequest").style.display = "none";
        document.getElementById("frontThemes").style.display = "none";
    }, true);

    document.getElementById("tabRequest").addEventListener('click', function(){
        document.getElementById("frontMain").style.display = "none";
        document.getElementById("frontRequest").style.display = "block";
        document.getElementById("frontThemes").style.display = "none";
    }, true);

    document.getElementById("tabTheme").addEventListener('click', function(){
        document.getElementById("frontMain").style.display = "none";
        document.getElementById("frontRequest").style.display = "none";
        document.getElementById("frontThemes").style.display = "block";
    }, true);
});
