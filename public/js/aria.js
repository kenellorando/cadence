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
        let i=1;
        data.forEach(function(song){
          console.log("RESULT " + i)
          console.log(song.title);
          console.log(song.artist);
          console.log(song.album);
          i++;
          console.log("============")
        })
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
