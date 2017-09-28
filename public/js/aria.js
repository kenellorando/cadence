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
      success: function(results) {
        console.log("Success");
        console.log(results);
        alert("success");
        alert(results);
      }
    });
  });
})
