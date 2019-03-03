$(document).ready(function () {  
    // Full library list load
    // The api request is a simple GET
    $('#getLibrary').click(function (e) {
        console.log("Requesting the full library listing...");
        // GET request to library API endpoint, expected JSON  
        $.ajax({
            // TODO here: possibly replace .ajax with shorthand jquery getJSON
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
                    table += "<div>No song data was returned in the library listing, that's weird./div>";
                }

                table += "</table>";
                // Put table into library HTML
                document.getElementById("library").innerHTML = table;

            },
            error: function () {
                console.log("Error retrieving full library listing.");
            }
        });
    });
});