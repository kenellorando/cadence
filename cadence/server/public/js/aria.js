var streamSrcURL = "";

$(document).ready(function() {
	getListenURL()
	getHistory()
	getNowPlayingMetadata()
	getNowPlayingAlbumArt()
	getVersion()
	connectRadioData()
	postSearch()
	postRequestID()
});

function getVersion() {
	$.ajax({
		type: "GET",
		url: "/api/version",
		dataType: "json",
		success: function (data) {
			document.getElementById("release").innerHTML = data.Version;
		},
		error: function () {
			document.getElementById("release").innerHTML = "(N/A)";
		},
	});
}

function getNowPlayingMetadata() {
	$.ajax({
		type: "GET",
		url: "/api/nowplaying/metadata",
		dataType: "json",
		success: function (data) {
			$("#song").text(data.Title);
			$("#artist").text(data.Artist);
		},
		error: function () {
			$("#song").text("-");
			$("#artist").text("-");
		},
	});
}

function getNowPlayingAlbumArt() {
	$.ajax({
		type: "GET",
		url: "/api/nowplaying/albumart",
		dataType: "json",
		success: function (data) {
			var nowPlayingArtwork = "data:image/jpeg;base64," + data.Picture;
			$("#artwork").attr("src", nowPlayingArtwork);
		},
		error: function () {
			$("#artwork").attr("src", "./static/blank.jpg");
		},
	});
}

function getListenURL() {
	$.ajax({
		type: "GET",
		url: "/api/listenurl",
		dataType: "json",
		success: function (data) {
			if (data.ListenURL == "-/-") {
				$("#status").html("Disconnected from server.");
			} else {
				streamSrcURL = location.protocol + "//" + data.ListenURL;
				document.getElementById("stream").src = streamSrcURL;
				$("#status").html(
					"Connected: <a href='" +
						streamSrcURL +
						"'>" +
						streamSrcURL +
						"</a>"
				);
			}
		},
		error: function () {
			document.getElementById("stream").src = "";
			$("#status").html("Disconnected from server.");
		},
	});
}

function getHistory() {
	$.ajax({
		type: 'GET',
		url: "/api/history",
		dataType: "json",
		success: function(data) {
			var table = "<table class='table is-striped' id='historyResults'>";
			if (data.length === 0) {
				document.getElementById("historyStatus").innerHTML = "No history available (yet).";
			} else {
				table += "<thead><tr><th>Ended</th><th>Artist</th><th>Title</th></tr></thead><tbody>"
				data.reverse().forEach(function(song) {
					var delta = Math.round((+(new Date()) - (new Date(String(song.Ended)))) / 1000);

					var minute = 60
					var hour = minute * 60
					var day = hour * 24

					var timeAgo;

					if (delta < 30) {
						timeAgo = 'just now';
					} else if (delta < minute) {
						timeAgo = delta + ' seconds ago';
					} else if (delta < 2 * minute) {
						timeAgo = 'a minute ago'
					} else if (delta < hour) {
						timeAgo = Math.floor(delta / minute) + ' minutes ago';
					} else if (Math.floor(delta / hour) == 1) {
						timeAgo = '1 hour ago'
					} else if (delta < day) {
						timeAgo = Math.floor(delta / hour) + ' hours ago';
					}

					table += "<tr><td>" + timeAgo + "</td><td>" + song.Artist + "</td><td>" + song.Title + "</td></tr>";
				})
				table += "</tbody>"		
				document.getElementById("historyStatus").innerHTML = "";
			}
			table += "</table>";
			document.getElementById("historyResults").innerHTML = table;
		},
		error: function() {	
			document.getElementById("historyStatus").innerHTML = "Error. Could not get history.";
		}
	});
}

function postSearch() {
	var data = {};
	data.search = $("#searchInput").val();
	$.ajax({
		type: "POST",
		url: "/api/search",
		contentType: "application/json",
		data: JSON.stringify(data),
		dataType: "json", // expects a json response
		success: function (data) {
			var table =
				"<table class='table is-striped is-hoverable' id='searchResults'>";
			if (data === null) {
				// if no results from search
				document.getElementById("requestStatus").innerHTML =
					"Results: 0";
				var input = $("#searchInput").val();
				input = input.replace(/</g, "&lt;").replace(/>/g, "&gt;"); // Encode < and >, for error when placed back into no-results message
			} else {
				document.getElementById("requestStatus").innerHTML =
					"Results: " + data.length;
				table +=
					"<thead><tr><th>Artist</th><th>Title</th><th>Availability</th></tr></thead><tbody>";
				data.forEach(function (song) {
					table +=
						"<tr><td>" +
						song.Artist +
						"</td><td>" +
						song.Title +
						"</td><td><button class='button is-small is-light requestButton' data-id='" +
						escape(song.ID) +
						"'>Request</button></td></tr>";
				});
				table += "</tbody>";
			}
			table += "</table>";
			document.getElementById("searchResults").innerHTML = table;
		},
		error: function () {
			document.getElementById("requestStatus").innerHTML =
				"Error. Could not execute search.";
		},
	});
}

function postRequestID() {
	$(document).on("click", ".requestButton", function (e) {
		var data = {};
		data.ID = unescape(this.dataset.id);
		$.ajax({
			type: "POST",
			url: "/api/request/id",
			contentType: "application/json",
			data: JSON.stringify(data),
			success: function () {
				document.getElementById("requestStatus").innerHTML =
					"Request accepted!";
			},
			error: function () {
				document.getElementById("requestStatus").innerHTML =
					"Sorry, your request was not accepted. You may be rate limited.";
			},
		});
	});
}

function connectRadioData() {
	let eventSource = new EventSource("/api/radiodata/sse");
	eventSource.onerror = function (event) {
		setTimeout(function () {
			connectRadioData();
		}, 5000);
	}
	eventSource.addEventListener("title", function(event) {
		$('#song').text(event.data)
	})
	eventSource.addEventListener("artist", function(event) {
		$('#artist').text(event.data)
	})
	eventSource.addEventListener("listeners", function(event) {
		if (event.data == -1) {
			$("#listeners").html("N/A");
		} else {
			$("#listeners").html(event.data);
		}
	})
	eventSource.addEventListener("title" || "artist" || "history", function() {
		getNowPlayingAlbumArt()
		getHistory()
	})
	eventSource.addEventListener("listenurl", function(event) {
		if (event.data == "-/-") {
			document.getElementById("stream").src = "";
			$("#status").html("Disconnected from server.");
		} else {
			streamSrcURL = location.protocol + "//" + event.data;
			document.getElementById("stream").src = streamSrcURL;
			$("#status").html(
				"Connected: <a href='" +
					streamSrcURL +
					"'>" +
					streamSrcURL +
					"</a>"
			);
		}
	});
}
