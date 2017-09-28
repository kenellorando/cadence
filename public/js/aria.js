/**
 * ARIA's Frontend Functionality
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
        console.log(data[0]);
        console.log(data[0].title);
      },
      error: function() {
        console.log("Failure");
        alert("failure");
      }
    });
  });
})
