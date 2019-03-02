// Switch the display of the front matter
$(document).ready(function () {
    document.getElementById("banner").addEventListener('click', function(){
        document.getElementById("frontMain").style.display = "block";
        document.getElementById("frontRequest").style.display = "none";
        document.getElementById("frontLibrary").style.display = "none";
    }, true);

    document.getElementById("navRequest").addEventListener('click', function(){
        document.getElementById("frontMain").style.display = "none";
        document.getElementById("frontRequest").style.display = "block";
        document.getElementById("frontLibrary").style.display = "none";
    }, true);

    document.getElementById("navLibrary").addEventListener('click', function(){
        document.getElementById("frontMain").style.display = "none";
        document.getElementById("frontRequest").style.display = "none";
        document.getElementById("frontLibrary").style.display = "block";
    }, true);
});
