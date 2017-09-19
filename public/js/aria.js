$(document).ready(function() {
    // Prevent page reloading on search
    $('#searchButton').click(function(e) {
        e.preventDefault();

        var data = {};
        data.search = $('#searchInput').val();
        
        // Set a key to 
        console.log(data.search); // railgun
        console.log(data);

        $.ajax({
            type: 'POST',
            url: 'http://localhost:8080/search',
            dataType: 'application/json',
            data: data,
        });
    });
})