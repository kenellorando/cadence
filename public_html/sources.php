<!--
Sources are written at the top here ONCE.
The /js/source-setter.js file takes these PHP variables and creates Javascript variables. Then the sources can be utilized and set into HTML and other Javascript pages as long as the source-setter is included first.
-->
<?php
  # Just the host IP
  $hostSource = '169.254.131.220';
  # Music Stream Server
  $streamSource = 'http://169.254.131.220:8000/cadence1';
  # Now Playing XSL sheet
  $nowPlayingSource = 'http://169.254.131.220:8000/now-playing.xsl';


  # Just the host IP
  echo ("hostSource: ");
  echo $hostSource;
/*
  function hostSource() {
    echo $hostSource();
  }
*/
  # Music Stream Server
  echo ("streamSource: ");
  echo $streamSource;
/*
  function streamSource() {
    echo $streamSource();
  }
  */
  # Just the host IP
  echo ("nowPlayingSource: ");
  echo $nowPlayingSource;
/*
  function nowPlayingSource() {
    echo $nowPlayingSource();
  }
*/

?>
