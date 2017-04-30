function selectChicagoEvening() {
  document.getElementById("selected-css").href = "/css/themes/chicago-evening.css";
  document.getElementById("title").innerHTML = "CADENCE";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Experience";
  localStorage.setItem('themeKey', 'chicagoEvening');
}

function selectCyberpunkBartender() {
  document.getElementById("selected-css").href = "/css/themes/cyberpunk-bartender.css";
  document.getElementById("title").innerHTML = "CADEN<span>C</span>E";
  document.getElementById("subtitle").innerHTML = "A Retro Cyberpunk Jukebox";
  localStorage.setItem('themeKey', 'cyberpunkBartender');
}

function selectMayberry() {
  document.getElementById("selected-css").href = "/css/themes/mayberry.css";
  document.getElementById("title").innerHTML = "CADENCE";
  document.getElementById("subtitle").innerHTML = "A Rhythmic Ξxperience";
  localStorage.setItem('themeKey', 'mayberry');
}

function selectElectromaster() {
  document.getElementById("selected-css").href = "/css/themes/electromaster.css";
  document.getElementById("title").innerHTML = "ケイデンス";
  document.getElementById("subtitle").innerHTML = "A Certain Scientific Radio";
  localStorage.setItem('themeKey', 'electromaster');
}

// This is run onload. To change the default theme, (for users that have not yet picked one) change the statement for null
function defaultTheme() {
  var theme = localStorage.getItem('themeKey');
  if (theme === "chicagoEvening") {
    selectChicagoEvening();
  } else if (theme === "cyberpunkBartender") {
    selectCyberpunkBartender();
  } else if (theme === "mayberry") {
    selectMayberry();
  } else if (theme === "electromaster") {
    selectElectromaster();
  } else if (theme === null) {
    selectChicagoEvening();
  }
}
