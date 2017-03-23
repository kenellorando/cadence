function selectChicagoEvening() {
  document.getElementById("selected-css").href = "/css/themes/chicago-evening.css";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Experience";
  document.getElementById("theme-tracker").value = themeID("chicago-evening");
}


function selectCyberpunkBartender() {
  document.getElementById("selected-css").href = "/css/themes/cyberpunk-bartender.css";
  document.getElementById("subtitle").innerHTML = "A Retro Cyberpunk Jukebox";
  document.getElementById("theme-tracker").value = themeID("cyberpunk-bartender");
}
