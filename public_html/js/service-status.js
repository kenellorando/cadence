jQuery(function ($) {
  $('#statusWebserverFTP').load('/status/index.php #statusWebserverFTP');
  $('#statusMusicStream').load('/status/index.php #statusMusicStream');
  $('#statusSongDatabase').load('/status/index.php #statusSongDatabase');
  $('#statusMusicStream2').load('/status/index.php #statusMusicStream2');
});

function checkStatus() {
  var statusTable = document.getElementById("statusTable");

  if (statusTable.style.display === 'none') {
    statusTable.style.display = 'block';
    document.getElementById("statusToggle").innerHTML = "Hide Service Status";
  } else {
    statusTable.style.display = 'none';
    document.getElementById("statusToggle").innerHTML = "Show Service Status";
  }
}
