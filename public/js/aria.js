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
        console.log("Success");
        console.log("=================")
        let i=1;
        if (data.length !== 0) {
          data.forEach(function(song){
            console.log("RESULT " + i)
            console.log("Title:" + song.title);
            console.log("Artist(s)" + song.artist);
            console.log("Album" + song.album);
            i++;
            console.log("=================")
          })
        } else {
          console.log("No results found. :(");
        }
        
        /*
        console.log(data[0]);
        console.log(data[0].title);
        */
      },
      error: function() {
        console.log("Failure");
      }
    });
  });
})
