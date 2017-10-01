/**
 * ARIA's Async Engine
 */

$(document).ready(function () {
  // Place responses from the server into this
  var ariaSays = document.getElementById("ariaSays");

  // Search button
  $('#searchButton').click(function (e) {
    // Create a key 'search' to send in JSON
    var data = {};
    data.search = $('#searchInput').val();

    $.ajax({
      type: 'POST',
      url: 'http://cadenceradio.com/search',
      dataType: 'application/json',
      data: data,
      dataType: "json",
      success: function (data) {
        console.log("Database query completed.");
        // console.log("=================");
        let i = 1;

        // Create the container table
        var table = "<table id = 'searchResults'>";

        if (data.length !== 0) {
          console.log("Cadence: Search completed. " + i + " result(s) found.")
          table += "<tr><th>Title</th><th>Artist</th><th>Availability</th></tr>"

          data.forEach(function (song) {
            /*
            console.log("RESULT " + i)
            console.log("Title: " + song.title);
            console.log("Artist(s): " + song.artist);
            console.log("Album: " + song.album);
            i++;
            console.log("=================");
            */
            table += "<tr><td class='dataTitle'>" + song.title + "</td><td class='dataArtist'>" + song.artist + "</td><td class='dataRequest'><button class='requestButton' data-path='" + song.path + "'>REQUEST</button></td></tr>";
          })
        } else {
          console.log("CADENCE: Search completed.  0 results found. :(");
        }

        table += "</table>";
        // Put table into results html
        document.getElementById("results").innerHTML = table;
      },
      error: function () {
        console.log("CADENCE: Error. Could not execute search.");
      }
    });
  });



  // Request buttons
  $(document).on('click', '.requestButton', function (e) {
    // console.log(this.dataset.path); // /home/ken/Music/fripSide/01. only my railgun.mp3

    var data = {};
    data.path = this.dataset.path;

    // Disable the request buttons for a certain amount of time
    $(".requestButton").prop('disabled', true);
    setTimeout(function() {
      $(".requestButton").prop('disabled', false);
    }, 5000);
    
    
    $.ajax({
      type: 'POST',
      url: 'http://cadenceradio.com/request',
      data: data,
      success: function (result) {
        console.log(result);
        ariaSays.innerHTML = result;
      },
      error: function (result) {
        console.log(result.responseText);
        ariaSays.innerHTML = result;
      }
    });
  });
});