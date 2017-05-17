/* From the /status/index.php page, grabs those PHP variables and sets them into Javascript variables
*/
jQuery(function ($) {
  var hostSource = load('/status/index.php #hostSource');
  var streamSource = load('/status/index.php #streamSource');
  var nowPlayingSource = load('/status/index.php #nowPlayingSource');
});
