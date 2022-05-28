$(document).ready(function() {
	// Initial search on load
	postSearch()
	// When the user presses the return key
	$("#searchInput").keyup(function(event) {
		if (event.keyCode == 13) {
			postSearch()
		}
	});
	// User clicks the search button
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

// Get latest source release title
$(document).ready(function() {
	$.ajax({
		type: 'GET',
		url: "/api/aria1/version",
		dataType: "json",
		// On success, format data into table
		success: function(data) {
			document.getElementById("release").innerHTML = data.Version;
		},
		error: function() {
			document.getElementById("release").innerHTML = "(N/A)";
		}
	});
});

var streamSrcURL = "" // this gets used by the stream playButton function
// Hook into the cadence radio data socket
$(document).ready(function() {
	if (location.protocol == "https:") {
		var socket = new WebSocket("wss://" +  location.host + "/api/aria1/radiodata/socket")
	} else {
		// This is necessary for local testing. All public ingress is https.
		var socket = new WebSocket("ws://" + location.host + "/api/aria1/radiodata/socket")
	}

	socket.onopen = () => {
		console.log("Established connection with Cadence radiodata socket.")
	}
	socket.onmessage = (ServerMessage) => {
		handle(ServerMessage)
	}
	socket.onerror = (ServerMessage) => {
		console.warn("Could not reach the Cadence radio data socket: " + ServerMessage.data)
	}

	function handle(ServerMessage) {
		let message = JSON.parse(ServerMessage.data)
		switch (message.Type) {
			case "NowPlaying":
				var nowPlayingArtist = message.Artist.trim();
				var nowPlayingTitle = message.Title.trim();
				$('#artist').text(nowPlayingArtist);
				$('#song').text(nowPlayingTitle);
				console.log("Now playing: " + nowPlayingArtist + ", '" + nowPlayingTitle + "'")
				break;
			case "Listeners":
				var currentListeners =  message.Listeners;
				if (currentListeners == -1) {
					document.getElementById("listeners").innerHTML = "(stream unreachable)"
				} else {
					document.getElementById("listeners").innerHTML = currentListeners;
				}
				break;
			case "StreamConnection":
				var currentListenURL = location.protocol + "//" + message.ListenURL.trim();
				var currentMountpoint = message.Mountpoint.trim();
				
				if (currentListenURL !== "N/A") {
					$('#status').html("Connected to stream: <a href='"+ currentListenURL + "'>" + currentMountpoint + "</a>");
				} else {
					$('#status').html("Disconnected from stream.");
				}
				document.getElementById("stream").src = currentListenURL
				streamSrcURL = currentListenURL // set global URL
				break;
		}
	}
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