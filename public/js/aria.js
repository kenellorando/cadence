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
      success: function(data) {
        document.getElementById("searchFailure").style.visibility="hidden";
        console.log("Success");
        console.log("=================")
        let i=1;
        var results=document.getElementById("results");
        if (data.length !== 0) {
          // Display results data
          results.style.display="block";
          // Remove all current results
          results.innerHTML="";

          data.forEach(function(song){
            // For each result, first log data
            console.log("RESULT " + i)
            console.log("Title: " + song.title);
            console.log("Artist(s): " + song.artist);
            console.log("Album: " + song.album);
            i++;
            console.log("=================")

            // Now add a result element to the results div
            var result=document.createElement("div");
            result.class="result";

            var artist=document.createElement("div");
            artist.class="artist_name";

            var album=document.createElement("div");
            album.class="album_title";

            var title=document.createElement("div");
            title.class="song_title";

            artist.innerHTML=song.artist;
            album.innerHTML=song.album;
            title.innerHTML=song.title;

            result.appendChild(artist);
            result.appendChild(album);
            result.appendChild(title);

            results.appendChild(result);
          })
        } else {
          console.log("No results found. :(");
          // Hide results div
          results.style.display="none"
        }
      },
      error: function() {
        console.log("Failure");
        document.getElementById("searchFailure").style.visibility="visible";
      }
    });
  });
})
