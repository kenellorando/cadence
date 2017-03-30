var cyberpunkCancel;

function cancelSwitcher() {
    clearTimeout(cyberpunkCancel);
    document.getElementsByTagName("html")[0].style.backgroundImage = "";
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

    document.getElementsByTagName("html")[0].style.backgroundImage = "url("+url+")";

    var maxTime = 1200; // ms
    var minTime = 100; // ms
    var timeRange = maxTime-minTime;
    var time = Math.floor(Math.random() * timeRange) + minTime;
    
    cancelSwitcher();
    
    cyberpunkCancel = setTimeout(cyberpunkBackgroundSwitcher, time);
    
    console.log(time);
}