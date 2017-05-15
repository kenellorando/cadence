jQuery(function ($) {
  $('#statusMusicStream').load('/status/index.php #statusMusicStream');
  $('#statusSongDatabase').load('/status/index.php #statusSongDatabase');
  $('#statusWebserverFTP').load('/status/index.php #statusWebserverFTP');
  
  // Tests
});

function show() {
  var old = document.getElementById("old");

  if (old.style.display === 'block') {
    old.style.display = 'none';
    document.getElementById("oldToggle").innerHTML = "Show Full History";
  } else {
    old.style.display = 'block';
    document.getElementById("oldToggle").innerHTML = "Hide Full History";
  }
}
