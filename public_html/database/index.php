<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="" />

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
  <link rel="stylesheet" href="/css/themes/base.css">
  <!-- Page theme -->
  <!--
  <link rel="stylesheet" href="/css/themes/your-name.css">
  -->
  <!-- jQuery Google CDN -->
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js"></script>
</head>

<body>
  <h1 id="title">CADENCE</h1>
  <p id="subtitle">A Name To Remember</p>

  <div id="content">
    <div class="content-left">
      <p>Connection to database: </p>
      <?php
        $servername = "localhost";
        // Query has permission only to select
        $username = "kenellor_query";
        $password = "query_pass";

        // Create connection
        $con = new mysqli($servername, $username, $password);

        // Check connection
        if ($con->connect_error) {
            die("Failed " . $con->connect_error);
        } else {
          echo "Success";
        }
      ?>
    </div>
    <div class="content-right">
      <form action="">
        <p>Search query: </p>
        <input type="text" name="search" />
        <input type="submit" value="Search" />
      </form>

      <?php
        
      ?>

    </div>
  </div>
  <footer></footer>
</body>

</html>
