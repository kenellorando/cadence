var streamSrcURL = "" // this is used by the stream source loader

$(window).on("load", function(e) {
	getRadioData()
	connectToSSE()
	postSearch()
	setSearchRequestFunctions()
});

function getRadioData() {
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
			} else {
				streamSrcURL = location.protocol + "//" + data.ListenURL
				document.getElementById("stream").src = streamSrcURL;
				$('#status').html("Connected: <a href='"+ streamSrcURL + "'>" + streamSrcURL + "</a>")
			}
		},
		error: function() {
			document.getElementById("stream").src = "";
			$('#status').html("Disconnected from server.")
		}
	});
}

var retry = 1;

function connectToSSE() {
	let eventSource = new EventSource("/api/radiodata/sse");
	window.addEventListener('beforeunload', () => {
		eventSource.close();
	});

	eventSource.onerror = function(event) {
		eventSource.close();
		retry *= 2;
		setTimeout(() => { connectToSSE(); }, retry * 1000);
	}

	eventSource.addEventListener("title", function(event) {
		$('#song').text(event.data)
	})
	eventSource.addEventListener("artist", function(event) {
		$('#artist').text(event.data)
	})
	eventSource.addEventListener("title" || "artist", function(event) {
		setAlbumArt()
	})
	eventSource.addEventListener("listeners", function(event) {
		if (event.data == -1) {
			$('#listeners').html("(stream unreachable)")
		} else {
			$('#listeners').html(event.data)
		}
	})
	eventSource.addEventListener("listenurl", function(event) {
		if (event.data == "-/-") {
			document.getElementById("stream").src = "";
			$('#status').html("Disconnected from server.")
		} else {
			streamSrcURL = location.protocol + "//" + event.data 
			document.getElementById("stream").src = streamSrcURL
			$('#status').html("Connected: <a href='"+ streamSrcURL + "'>" + streamSrcURL + "</a>")
		}
	})
}

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
			var nowPlayingArtwork = "data:image/jpeg;base64,"+data.Picture;
			$('#artwork').attr("src", nowPlayingArtwork);
		},
		error: function() {
			$('#artwork').attr("src", "");
		}
	});
}

function setSearchRequestFunctions() {
	// Presses return key
	$("#searchInput").keyup(function(event) {
		if (event.keyCode == 13) {
			postSearch()
		}
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
			success: function() {
				document.getElementById("requestStatus").innerHTML = "Request accepted!";
			},
			error: function() {
				document.getElementById("requestStatus").innerHTML = "Sorry, your request is rate limited.";
			},
		})
	})
}
