$(document).ready(function () {  
    // Full library list load
    // This is a GET request to /api/aria1/library
    $('#getLibrary').click(function (e) {
        console.log("Requesting the full library listing...");
        document.getElementById("library").innerHTML = "<div>Getting full library listing...</div>";

        // GET request to library API endpoint, expected JSON  
        $.ajax({
            type: 'GET',
            url: '/api/aria1/library',
            dataType: 'json',
            // On success, format data into table
            success: function (data) {
                console.log("Successfully retrieved full library listing.")
                console.log(data)
    
                
                // Start the containing table
                let table = "<table id='libraryTable'>";
                let i = 1;

                if (data.length !== 0) {
                    table += "<tr><th>Artist</th><th>Title</th></tr>"

                    data.forEach(function (song) {
                        table += "<tr><td>" + song.Artist + "</td><td>" + song.Title + "</td></tr>";
                    })
                } else {
                    document.getElementById("library").innerHTML = "<div>Couldn't get full library listing! :(</div>";
                }

                table += "</table>";
                // Put table into library HTML
                document.getElementById("library").innerHTML = table;
            },
            error: function () {
                console.log("Error retrieving full library listing.");
                document.getElementById("library").innerHTML = "<div>Couldn't get full library listing! :(</div>";
            }
        });
    });


    // Search box under request tab, handles when the user presses 'enter'
    $("#searchInput").keyup(function (event) {
        // Keycode 13 is the return key.
        if (event.keyCode == 13) {
            // Simply simulate a click on the search button itself
            $("#searchButton").click();
        }
    });
    // Search box under request tab, handles when the user clicks the search button
    $('#searchButton').click(function (e) {
        // Create a key 'search' to send in JSON
        var data = {};
        data.search = $('#searchInput').val();

        $.ajax({
            type: 'POST',
            url: '/api/aria1/search',
            /* contentType sends application/x-www-form-urlencoded data */
            contentType: 'application/x-www-form-urlencoded',
            data: JSON.stringify(data),
            /* dataType expects a json response */
            dataType: 'json',
            success: function (data) {
                let i = 1;

                // Create the container table
                var table = "<table id = 'searchResults'>";

                if (data === null) {
                    console.log("Search completed.  0 results found.");
                    document.getElementById("requestStatus").innerHTML = "Search completed.  0 results found.";

                    // Encode < and >, for error when placed back into no-results message
                    var input = $('#searchInput').val();
                    input = input.replace(/</g, "&lt;").replace(/>/g, "&gt;");
                    // No-results message
                    table += "<div>Nothing found for search '"+input+"' :(</div>";
                } else {
                    console.log("Search completed. Results found: " + data.length)
                    document.getElementById("requestStatus").innerHTML = "Search completed. Results found: " + data.length;

                    // Build the results table
                    table += "<tr><th>Artist</th><th>Title</th><th>Availability</th></tr>"
                    data.forEach(function (song) {
                        table += "<tr><td>" + song.Artist + "</td><td>" + song.Title + "</td><td><button class='requestButton' data-id='" + escape(song.ID) + "'>REQUEST</button></td></tr>";
                    })
                }

                table += "</table>";
                // Put table into results html
                document.getElementById("searchResults").innerHTML = table;
            },
            error: function () {
                console.log("Error. Could not execute search.");
            }
        });
    });


    // Clicks on song request buttons
    $(document).on('click', '.requestButton', function (e) {
        var data = {};
        data.ID = unescape(this.dataset.id);

        $.ajax({
            type: 'POST',
            url: '/api/aria1/request',
            /* contentType sends application/x-www-form-urlencoded data */
            contentType: 'application/x-www-form-urlencoded',
            data: JSON.stringify(data),
            /* dataType expects a text response */
            dataType: 'text',
            success: function (data) {
                console.log("Song request submitted.");
                document.getElementById("requestStatus").innerHTML = "Request submitted!";
            },
            error: function () {
                console.log("Error. Something went wrong submitting the song request..");
                document.getElementById("requestStatus").innerHTML = "Error. Something went wrong submitting the song request..";
            }
        });
    })
});