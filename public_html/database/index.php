<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="" />

  <title>Cadence Database</title>
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
</head>

<body>
  <?php

    /* Attempt MySQL server connection. Assuming you are running MySQL server with default setting (user 'root' with no password) */

    $link = mysqli_connect("localhost", "kenellor_query", "query1", "kenellor_cadence");


    /* Check connection */
    if($link === false){
        die("ERROR: Could not connect. " . mysqli_connect_error());
    }
     
    /* Attempt select query execution */
    $sql = "SELECT * FROM music";
    if($result = mysqli_query($link, $sql)){
        if(mysqli_num_rows($result) > 0){
            echo "<table>";
                echo "<tr>";
                    echo "<th>song_id</th>";
                    echo "<th>song_title</th>";
                    echo "<th>song_path</th>";
                echo "</tr>";
            while($row = mysqli_fetch_array($result)){
                echo "<tr>";
                    echo "<td>" . $row['song_id'] . "</td>";
                    echo "<td>" . $row['song_title'] . "</td>";
                    echo "<td>" . $row['song_path'] . "</td>";
                echo "</tr>";
            }
            echo "</table>";
            mysqli_free_result($result);
        } else{
            echo "No records matching your query were found.";
        }
    } else{
        echo "ERROR: Could not able to execute $sql. " . mysqli_error($link);
    }
    /* Close connection */
    mysqli_close($link);
    ?>
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
