<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="Cadence Radio - An open source, commercial-free radio by Ken Ellorando " />
  <meta name="keywords" content="Cadence, Radio, Cadence Radio, CadenceRadio, free radio, open source radio, github, Ken Ellorando radio" />

  <title>Cadence Radio</title>
  <link rel="shortcut icon" href="favicon.ico" type="image/x-icon">

  <!-- OLD FONT: Permenant Marker
	Heading: Rock Salt; Subtitle: Roboto 300; All else: PT Sans -->
  <link href="https://fonts.googleapis.com/css?family=Rock+Salt" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css?family=Roboto:300i" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css?family=PT+Sans" rel="stylesheet">

  <!-- Normalization CSS -->
  <link rel="stylesheet" href="/css/normalize.css">
  <!-- BASE CSS -->
  <link rel="stylesheet" id="base-css" href="/css/themes/base.css">
  <!-- Selected Style CSS -->
  <link rel="stylesheet" id="selected-css" href="">
  <!-- Status CSS -->
  <link rel="stylesheet" href="/css/status/status.css">

  <!-- jQuery Google CDN -->
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js"></script>
  <!-- Theme changes after base -->
  <script src="/js/theme-changer.js"></script>
  <!-- Clock -->
  <script src="/js/clock.js"></script>
</head>


<body onload="defaultTheme(); clock();">
  <h1>Cadence Radio Live Status</h1>
  <!-- Clock -->
  <div id="heading-time">
    <div>Server Time:
      <?php
      date_default_timezone_set('America/Chicago');

      $timestamp = time();
      $date_time = date("d-m-Y (D) H:i:s", $timestamp);
      echo "$date_time";
      ?> (UTC-6)
    </div>
    <div>Local Time: <span id="clock"></span></div>
  </div>

  <ul>
    <li>Music Server Status:
      <?php
          $host = 'http://169.254.131.220'; 
            $port = 8000; 
            $waitTimeoutInSeconds = 3; 
            if($fp = fsockopen($host,$port,$errCode,$errStr,$waitTimeoutInSeconds)){   
               echo ("<div style='color:#7CFC00'> ONLINE </div>");
            } else {
               echo ("<div style ='color:#cc0000'> OFFLINE </div>");
            } 
            fclose($fp);
        ?>
    </li>
    <li>Song Metadata Database Status:
      <?php
        $host = 'http://169.254.131.220'; 
          $port = 3306; 
          $waitTimeoutInSeconds = 3; 
          if($fp = fsockopen($host,$port,$errCode,$errStr,$waitTimeoutInSeconds)){   
             echo ("<div style='color:#7CFC00'> ONLINE </div>");
            } else {
               echo ("<div style ='color:#cc0000'> OFFLINE </div>");
          } 
          fclose($fp);
      ?>
    </li>
  </ul>
</body>

</html>
