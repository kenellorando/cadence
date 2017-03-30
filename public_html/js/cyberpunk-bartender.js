var cyberpunkCancel;

function cancelSwitcher() {
    clearTimeout(cyberpunkCancel);
    document.getElementsByTagName("html")[0].style.backgroundImage = "";
}

function fadeFromWhite() {
    var html = document.getElementsByTagName("html")[0];

    html.classList.add("transition");

    setTimeout(function(){
        html.classList.remove("transition");

        cyberpunkBackgroundSwitcher();
    }, 3000);
}

// Handles the switching of the cyberpunk bartender background
function cyberpunkBackgroundSwitcher() {
    var URLs = [
    "cyberpunk-bartender.gif",
    "cyberpunk-bartender2.gif",
    "cyberpunk-bartender3.gif"];
    var URLroot = document.location.href;
    URLroot = URLroot.substring(0, URLroot.lastIndexOf("/")+1);
    URLroot += "media/";

    var index = Math.floor(Math.random() * URLs.length);
    
    var url = URLroot+URLs[index]

    var maxTime = 120000; // ms
    var minTime = 10000; // ms
    var timeRange = maxTime-minTime;
    var time = Math.floor(Math.random() * timeRange) + minTime;
    
    cancelSwitcher();

    document.getElementsByTagName("html")[0].style.backgroundImage = "url("+url+")";
    
    cyberpunkCancel = setTimeout(fadeFromWhite, time);
}