<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="Cadence Radio - An open source, commercial-free radio by Ken Ellorando " />
  <meta name="keywords" content="Cadence, Radio, Cadence Radio, CadenceRadio, free radio, open source radio, github, Ken Ellorando radio" />

  <title>Cadence Status</title>
  <link rel="shortcut icon" href="favicon.ico" type="image/x-icon">

  <!-- OLD FONT: Permenant Marker
	Heading: Rock Salt; Subtitle: Roboto 300; All else: PT Sans -->
  <link href="https://fonts.googleapis.com/css?family=Rock+Salt" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css?family=Roboto:300i" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css?family=PT+Sans" rel="stylesheet">

  <!-- Normalization CSS -->
  <link rel="stylesheet" href="/css/normalize.css">

  <!-- jQuery Google CDN -->
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js"></script>
  <!-- Clock -->
  <script src="/js/clock.js"></script>
</head>


<body onload="clock();">
  <h1>Cadence Radio Status</h1>
  <!-- Clock -->
  <div id="heading-time">
    <div>Data as of Server Time:
      <?php
      date_default_timezone_set('America/Chicago');

      $timestamp = time();
      //$date_time = date("d-m-Y (D) H:i:s", $timestamp);
      $date_time = date("H:i:s", $timestamp);
      echo "$date_time";
      ?> (UTC-6)
    </div>
    <div>Local Time: <span id="clock"></span></div>
  </div>

  <ul>
    <!-- Primary Stream -->
    <li>
      <div id="statusMusicStream">
        <?php
          $host = '169.254.131.220'; 
            $port = 8000; 
            $waitTimeoutInSeconds = 2; 
            if($fp = fsockopen($host,$port,$errCode,$errStr,$waitTimeoutInSeconds)){   
               echo ("<div style='color:#7CFC00'> ONLINE </div>");
            } else {
               echo ("<div style ='color:#cc0000'> OFFLINE </div>");
            } 
            fclose($fp);
        ?>
      </div>
    </li>
    <!-- Metadata Database -->
    <li>
      <div id="statusSongDatabase">
        <?php
        $host = 'localhost'; 
          $port = 2083; 
          $waitTimeoutInSeconds = 2; 
          if($fp = fsockopen($host,$port,$errCode,$errStr,$waitTimeoutInSeconds)){ 
            
             $servername = "localhost";
        // Query has permission only to select
        $username = "kenellor_query";
        $password = "query1";

        // Create connection
        $con = new mysqli($servername, $username, $password);

        // Check connection
        if ($con->connect_error) {
          // die("Failed " . $con->connect_error);
          echo("<div style ='color:#FF6347'> ONLINE  (QUERYING FAILED) </div>");
        } else {
          echo ("<div style='color:#7CFC00'> ONLINE </div>");
        }
            } else {
               echo ("<div style ='color:#cc0000'> OFFLINE </div>");
          } 
          fclose($fp);
      ?>
      </div>
    </li>
    <!-- Webserver FTP -->
    <li>
      <div id="statusWebserverFTP">
        <?php
          $host = 'localhost'; 
            $port = 21; 
            $waitTimeoutInSeconds = 2; 
            if($fp = fsockopen($host,$port,$errCode,$errStr,$waitTimeoutInSeconds)){   
               echo ("<div style='color:#7CFC00'> ONLINE </div>");
            } else {
               echo ("<div style ='color:#cc0000'> OFFLINE </div>");
            } 
            fclose($fp);
        ?>
      </div>
    </li>
  </ul>
</body>

</html>
