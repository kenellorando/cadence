// Switch the display of the front matter
$(document).ready(function () {
    // Load this template for the navigation bar
    document.getElementById("nav").innerHTML = `
<span id="navLeft">
    <a id="banner">CADENCE</a>
</span>
<span id="navRight">
    <a id="navRequest">Request</a>
    <a id="navLibrary">Library</a>
    <a href="https://github.com/kenellorando/cadence" target="_blank">Source</a>
</span>
`

    // Load this template for the page footer
    document.getElementById("footer").innerHTML = `
<span>A project by <a href="https://kenellorando.com/">Ken Ellorando</a></span>
`

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
