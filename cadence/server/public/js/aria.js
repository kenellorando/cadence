$(window).on("load", function(e) {
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
			url: '/api/request/id',
			/* contentType sends application/x-www-form-urlencoded data */
			contentType: 'application/x-www-form-urlencoded',
			data: JSON.stringify(data),
			/* dataType expects a json response */
			dataType: 'json',
			complete: function(data) {
				// console.log("Server message: " + data.responseJSON.Message);
				// console.log("Timeout remaining (s): " + data.responseJSON.TimeRemaining);
				
				document.getElementById("requestStatus").innerHTML = "Request submitted!";
				// // Disable the request button over UI
				// $(".requestButton").prop('disabled', true);
				// document.getElementById("moduleRequestButton").href = "/css/modules/requestButtonDisabled.css"
				// // Enable the request button after X minutes
				// setTimeout(function() {
				// 	$(".requestButton").prop('disabled', false);
				// 	document.getElementById("moduleRequestButton").href = "/css/modules/requestButtonEnabled.css"
				// }, 1000 * data.responseJSON.TimeRemaining)
			}
		})
	})
});

// Initial data load
$(window).on("load", function(e) {
	$.ajax({
		type: 'GET',
		url: "/api/version",
		dataType: "json",
		success: function(data) {
			document.getElementById("release").innerHTML = data.Version;
		},
		error: function() {
			document.getElementById("release").innerHTML = "(N/A)";
		}
	});
	$.ajax({
		type: 'GET',
		url: "/api/nowplaying/metadata",
		dataType: "json",
		success: function(data) {
			$('#song').text(data.Title)
			$('#artist').text(data.Artist)
		},
		error: function() {
			$('#song').text("-")
			$('#artist').text("-")
		}
	});
	$.ajax({
		type: 'GET',
		url: "/api/nowplaying/albumart",
		dataType: "json",
		success: function(data) {
			$('#artwork').attr("src", "data:image/jpeg;base64," + data.Picture);
		},
		error: function() {
			$('#artwork').attr("src", "");
		}
	});
	$.ajax({
		type: 'GET',
		url: "/api/listenurl",
		dataType: "json",
		success: function(data) {
			if (data.ListenURL == "-/-") {
				$('#status').html("Disconnected from server.")
				$.ajax(this);
				return
			} else {
				streamSrcURL = location.protocol + "//" + data.ListenURL
				document.getElementById("stream").src = streamSrcURL;
				$('#status').html("Connected: <a href='"+ streamSrcURL + "'>" + streamSrcURL + "</a>")
			}
		},
		error: function() {
			document.getElementById("stream").src = "";
			$.ajax(this);
			return
		}
	});
});

var streamSrcURL = "" // this gets used by the stream playButton function

// Check into the cadence nowplaying metadata event stream
$(document).ready(function() {
	let eventSource = new EventSource("/api/radiodata/sse");
	eventSource.onopen = function(event) {
		console.log("connected", event);
	}
	eventSource.onerror = function(event) {
		console.log("error connecting", event);
	}
	eventSource.addEventListener("title", function(event) {
		$('#song').text(event.data)
		setAlbumArt()
	})
	eventSource.addEventListener("artist", function(event) {
		$('#artist').text(event.data)
	})
	eventSource.addEventListener("listeners", function(event) {
		let listenerUpdate = event.data
		if (listenerUpdate == -1) {
			$('#listeners').html("(stream unreachable)")
		} else {
			$('#listeners').html(listenerUpdate)
		}
	})
	eventSource.addEventListener("listenurl", function(event) {
		let listenurl = event.data
		if (listenurl == "-/-") {
			$('#status').html("Disconnected from server.")
		} else {
			streamSrcURL = location.protocol + "//" + listenurl 
			document.getElementById("stream").src = newListenURL
			$('#status').html("Connected: <a href='"+ newListenURL + "'>" + newListenURL + "</a>")
		}
	})
});

function postSearch() {
	// Create a key 'search' to send in JSON
	var data = {};
	data.search = $('#searchInput').val();
	$.ajax({
		type: 'POST',
		url: '/api/search',
		/* contentType sends application/x-www-form-urlencoded data */
		contentType: 'application/x-www-form-urlencoded',
		data: JSON.stringify(data),
		/* dataType expects a json response */
		dataType: 'json',
		success: function(data) {
			let i = 1;
			// Create the container table
			var table = "<table class='table is-striped is-hoverable' id='searchResults'>";
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
				table += "<thead><tr><th>Artist</th><th>Title</th><th>Availability</th></tr></thead><tbody>"
				data.forEach(function(song) {
					table += "<tr><td>" + song.Artist + "</td><td>" + song.Title + "</td><td><button class='button requestButton' data-id='" + escape(song.ID) + "'>REQUEST</button></td></tr>";
				})
				table += "</tbody>"

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


// Get currently playing album art
function setAlbumArt() {
	$.ajax({
		type: 'GET',
		url: "/api/nowplaying/albumart",
		dataType: "json",
		// On success, switch the source of the artwork tag
		success: function(data) {
			var nowPlayingArtwork = data.Picture;
			$('#artwork').attr("src", "data:image/jpeg;base64,"+ nowPlayingArtwork);
		},
		error: function() {
			$('#artwork').attr("src", "");
		}
	});
}
