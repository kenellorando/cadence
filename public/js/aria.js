$(document).ready(function () {  
    // Full library list load
    // The api request is a simple GET
    $('#getLibrary').click(function (e) {
        console.log("Requesting the full library listing...");
        // GET request to library API endpoint, expected JSON  
        $.ajax({
            type: 'GET',
            url: '/api/aria1/library',
            dataType: 'application/json',
            // On success, format data into table
            success: function (data) {
                /*
                // Start the containing table
                let table = "<table id='libraryTable'>";
                let i = 1;

                if (data.length !== 0) {
                console.log("CADENCE: Database query completed. " + data.length + " result(s) found.")
                table += "<tr><th>Title</th><th>Artist</th><th>Availability</th></tr>"

                data.forEach(function (song) {
                    table += "<tr><td class='dataTitle'>" + song.title + "</td><td class='dataArtist'>" + song.artist + "</td><td class='dataRequest'><button class='requestButton' data-path='" + escape(song.path) + "'>REQUEST</button></td></tr>";
                })
                } else {
                console.log("CADENCE: Database query completed.  0 results found. :(");
                table += "<div style='padding-top: 2em'>Nothing found for search '"+input+"' :(</div>";
                }

                table += "</table>";
                // Put table into results html
                document.getElementById("library").innerHTML = table;
                */
               console.log("Success.")
               console.log(data.length)
               console.log(data)
            },
            error: function () {
                console.log("Error retrieving full library listing.");
            }
        });
    });
});