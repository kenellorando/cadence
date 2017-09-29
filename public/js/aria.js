/**
 * ARIA's Async Engine
 */
$(document).ready(function () {
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
        console.log("Success");
        console.log("=================")
        let i = 1;

        // Create the container table
        var table = document.createElement("table");

        if (data.length !== 0) {
          data.forEach(function (song) {
            console.log("RESULT " + i)
            console.log("Title: " + song.title);
            console.log("Artist(s): " + song.artist);
            console.log("Album: " + song.album);
            i++;
            console.log("=================")

            var resultsDiv = document.getElementById('results');

            // Row for this song
            var tableRow = document.createElement("tr");

            var songTitleData = document.createElement("td");
            var songArtistData = document.createElement("td");
            var songAlbumData = document.createElement("td");

            // Set the data
            songTitleData.innerHTML = song.title;
            songArtistData.innerHTML = song.artist;
            songAlbumData.innerHTML = song.album;
            // Append to the row
            tableRow.appendChild(songTitleData);
            tableRow.appendChild(songArtistData);
            tableRow.appendChild(songAlbumData);

            // Put row into table
            table.appendChild(tableRow);
          })
        } else {
          console.log("No results found. :(");
        }

        // Put table into results html
        document.getElementById("results").innerHTML = table;

      },
      error: function () {
        console.log("Failure");
      }
    });
  });
})