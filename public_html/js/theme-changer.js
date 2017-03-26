
function selectChicagoEvening() {
  document.getElementById("selected-css").href = "/css/themes/chicago-evening.css";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Experience";
  localStorage.setItem('themeKey', 'chicagoEvening');
}

function selectCyberpunkBartender() {
  document.getElementById("selected-css").href = "/css/themes/cyberpunk-bartender.css";
  document.getElementById("subtitle").innerHTML = "A Retro Cyberpunk Jukebox";
  localStorage.setItem('themeKey', 'cyberpunkBartender');
}

function defaultTheme() {
  var theme = localStorage.getItem('themeKey');
  if (theme === "chicagoEvening") {
    selectChicagoEvening();
  } else if (theme === "cyberpunkBartender") {
    selectCyberpunkBartender()
  }
}
