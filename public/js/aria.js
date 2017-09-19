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
      url: 'http://localhost:8080/search',
      dataType: 'application/json',
      data: data,
    });
  });
})
