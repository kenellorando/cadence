$(document).ready(function() {
	// Initial search on load
	postSearch()
	// Handles when the user keys "Return"
	$("#searchInput").keyup(function(event) {
		$("#searchButton").click();
	});
	// Handles when the user clicks the search button
	$('#searchButton').click(function(e) {
		postSearch()
	});

	// Clicks on song request buttons
	$(document).on('click', '.requestButton', function(e) {
		var data = {};
		data.ID = unescape(this.dataset.id);
		$.ajax({
			type: 'POST',
			url: '/api/aria1/request',
			/* contentType sends application/x-www-form-urlencoded data */
			contentType: 'application/x-www-form-urlencoded',
			data: JSON.stringify(data),
			/* dataType expects a json response */
			dataType: 'json',
			complete: function(data) {
				console.log("Server message: " + data.responseJSON.Message);
				console.log("Timeout remaining (s): " + data.responseJSON.TimeRemaining);
				document.getElementById("requestStatus").innerHTML = "Server message: " + data.responseJSON.Message;
				// Disable the request button over UI
				$(".requestButton").prop('disabled', true);
				document.getElementById("moduleRequestButton").href = "/css/modules/requestButtonDisabled.css"
				// Enable the request button after X minutes
				setTimeout(function() {
					$(".requestButton").prop('disabled', false);
					document.getElementById("moduleRequestButton").href = "/css/modules/requestButtonEnabled.css"
				}, 1000 * data.responseJSON.TimeRemaining)
			}
		})
	})
});


function postSearch() {
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
		success: function(data) {
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
				table += "<div>Nothing found for search '" + input + "' :(</div>";
			} else {
				console.log("Search completed. Results found: " + data.length)
				document.getElementById("requestStatus").innerHTML = "Search completed. Results found: " + data.length;
				// Build the results table
				table += "<tr><th>Artist</th><th>Title</th><th>Availability</th></tr>"
				data.forEach(function(song) {
					table += "<tr><td>" + song.Artist + "</td><td>" + song.Title + "</td><td><button class='requestButton' data-id='" + escape(song.ID) + "'>REQUEST</button></td></tr>";
				})
			}
			table += "</table>";
			// Put table into results html
			document.getElementById("searchResults").innerHTML = table;
		},
		error: function() {
			console.log("Error. Could not execute search.");
		}
	});
}