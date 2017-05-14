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
  <!-- Status CSS -->
  <link rel="stylesheet" id="selected-css" href="/css/status/status.css">

  <!-- jQuery Google CDN -->
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js"></script>
  <!-- Clock -->
  <script src="/js/clock.js"></script>
</head>


<body onload="clock();">
  <!-- Clock -->
  <div id="clock"></div>

  <ul>
    <li>Cadence Music Server Status:
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
    <li>Cadence Database Status
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
