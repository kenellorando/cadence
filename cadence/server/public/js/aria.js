
var streamSrcURL = ""

$(document).ready(function() {
	getListenURL()
	getNowPlayingMetadata()
	getNowPlayingAlbumArt()
	getVersion()
	connectRadioData()
	postSearch()
	postRequestID()

	function getVersion() {
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
	}

	function getNowPlayingMetadata() {
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
	}

	function getNowPlayingAlbumArt() {
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

	function getListenURL() {
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

	function postSearch() {
		var data = {};
		data.search = $('#searchInput').val();
		$.ajax({
			type: 'POST',
			url: '/api/search',
			contentType: 'application/x-www-form-urlencoded', // sends application/x-www-form-urlencoded data
			data: JSON.stringify(data),
			dataType: 'json', // expects a json response
			success: function(data) {
				var table = "<table class='table is-striped is-hoverable' id='searchResults'>";
				if (data === null) { // if no results from search
					document.getElementById("requestStatus").innerHTML = "Search completed.  0 results found.";
					var input = $('#searchInput').val();
					input = input.replace(/</g, "&lt;").replace(/>/g, "&gt;"); // Encode < and >, for error when placed back into no-results message
					table += "<div>Nothing found for search '" + input + "' :(</div>";
				} else {
					document.getElementById("requestStatus").innerHTML = "Search completed. Results found: " + data.length;
					table += "<thead><tr><th>Artist</th><th>Title</th><th>Availability</th></tr></thead><tbody>"
					data.forEach(function(song) {
						table += "<tr><td>" + song.Artist + "</td><td>" + song.Title + "</td><td><button class='button is-small is-light requestButton' data-id='" + escape(song.ID) + "'>Request</button></td></tr>";
					})
					table += "</tbody>"
				}
				table += "</table>";
				document.getElementById("searchResults").innerHTML = table;
			},
			error: function() {				
				document.getElementById("requestStatus").innerHTML = "Error. Could not execute search.";
			}
		});
	}

	function postRequestID() {
		$(document).on('click', '.requestButton', function(e) {
			var data = {};
			data.ID = unescape(this.dataset.id);
			$.ajax({
				type: 'POST',
				url: '/api/request/id',
				contentType: 'application/x-www-form-urlencoded', // sends application/x-www-form-urlencoded data
				data: JSON.stringify(data),
				success: function() {
					document.getElementById("requestStatus").innerHTML = "Request accepted!";
				},
				error: function() {
					document.getElementById("requestStatus").innerHTML = "Sorry, your request was not accepted. You may be rate limited.";
				},
			})
		})
	}

	function connectRadioData() {
		let eventSource = new EventSource("/api/radiodata/sse");
		eventSource.onerror = function(event) {
			setTimeout(function() { 
				connectRadioData(); 
			}, 5000);
		}
		eventSource.addEventListener("title", function(event) {
			$('#song').text(event.data)
		})
		eventSource.addEventListener("artist", function(event) {
			$('#artist').text(event.data)
		})
		eventSource.addEventListener("title" || "artist", function(event) {
			getNowPlayingAlbumArt()
		})
		eventSource.addEventListener("listeners", function(event) {
			if (event.data == -1) {
				$('#listeners').html("N/A")
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
});
