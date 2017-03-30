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
}