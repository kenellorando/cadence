const SONG_ID = '#song'
const ARTIST_ID = '#artist'
const ARTWORK_ID = '#artwork'
const STATUS_ID = '#status'
const SEARCH_INPUT_ID = '#searchInput'
const LISTENERS_ID = '#listeners'
const RELEASE_ID = 'release'
const STREAM_ID = 'stream'
const REQUEST_STATUS_ID = 'requestStatus'

let streamSrcURL = ''

$(document).ready(() => {
   getListenURL()
   getNowPlayingMetadata()
   getNowPlayingAlbumArt()
   getVersion()
   connectRadioData()
   postSearch()
   postRequestID()
})

let getVersion = () => {
   $.ajax({
      type: 'GET',
      url: '/api/version',
      dataType: 'json',
      success: (data) => document.getElementById(RELEASE_ID).innerHTML = data.Version,
      error: () => document.getElementById(RELEASE_ID).innerHTML = '(N/A)',
   })
}

const getNowPlayingMetadata = () => {
   $.ajax({
      type: 'GET',
      url: '/api/nowplaying/metadata',
      dataType: 'json',
      success: (data) => {
         $(SONG_ID).text(data.Title)
         $(ARTIST_ID).text(data.Artist)
      },
      error: () => {
         $(SONG_ID).text('-')
         $(ARTIST_ID).text('-')
      },
   })
}

const getNowPlayingAlbumArt = () => {
   $.ajax({
      type: 'GET',
      url: '/api/nowplaying/albumart',
      dataType: 'json',
      success: (data) => {
         let nowPlayingArtwork = `data:image/jpegbase64,${data.Picture}`
         $(ARTWORK_ID).attr('src', nowPlayingArtwork)
      },
      error: () => $(ARTWORK_ID).attr('src', ''),
   })
}

const getListenURL = () => {
   $.ajax({
      type: 'GET',
      url: '/api/listenurl',
      dataType: 'json',
      success: (data) => {
         if (data.ListenURL == '-/-') {
            $(STATUS_ID).html('Disconnected from server.')
         } else {
            streamSrcURL = `${location.protocol}//${data.ListenURL}`
            document.getElementById(STREAM_ID).src = streamSrcURL
            $(STATUS_ID).html(`Connected: <a href='${streamSrcURL}'>${streamSrcURL}</a>`)
         }
      },
      error: () => {
         document.getElementById(STREAM_ID).src = ''
         $(STATUS_ID).html('Disconnected from server.')
      },
   })
}

const postSearch = () => {
   let data = {}
   data.search = $(SEARCH_INPUT_ID).val()
   $.ajax({
      type: 'POST',
      url: '/api/search',
      contentType: 'application/json',
      data: JSON.stringify(data),
      dataType: 'json', // expects a json response
      success: (data) => {
         let table = "<table class='table is-striped is-hoverable' id='searchResults'>"

         if (!data) {
            // if no results from search
            document.getElementById(REQUEST_STATUS_ID).innerHTML = 'Results: 0'
            let input = $(SEARCH_INPUT_ID).val()
            input = input.replace(/</g, '&lt').replace(/>/g, '&gt') // Encode < and >, for error when placed back into no-results message
         } else {
            document.getElementById(REQUEST_STATUS_ID).innerHTML = `Results: ${data.length}`
            table += '<thead><tr><th>Artist</th><th>Title</th><th>Availability</th></tr></thead><tbody>'
            data.forEach((song) => table += `<tr><td>${song.Artist}</td><td>${song.Title}</td><td><button class='button is-small is-light requestButton' data-id='${decodeURI(song.ID)}'>Request</button></td></tr>`)
            table += '</tbody>'
         }
         table += '</table>'
         document.getElementById('searchResults').innerHTML = table
      },
      error: () => document.getElementById(REQUEST_STATUS_ID).innerHTML = 'Error. Could not execute search.',
   })
}

const postRequestID = () => {
   $(document).on('click', '.requestButton', (e) => {
      let data = {}
      data.ID = decodeURI(this.dataset.id)
      $.ajax({
         type: 'POST',
         url: '/api/request/id',
         contentType: 'application/json',
         data: JSON.stringify(data),
         success: () => document.getElementById(REQUEST_STATUS_ID).innerHTML = 'Request accepted!',
         error: () => document.getElementById(REQUEST_STATUS_ID).innerHTML = 'Sorry, your request was not accepted. You may be rate limited.',
      })
   })
}

const connectRadioData = () => {
   let eventSource = new EventSource('/api/radiodata/sse')

   eventSource.onerror = (event) => setTimeout(() => connectRadioData(), 5000)
   eventSource.addEventListener('title', (event) => $(SONG_ID).text(event.data))
   eventSource.addEventListener('artist', (event) => $(ARTIST_ID).text(event.data))
   eventSource.addEventListener('title' || 'artist', () => getNowPlayingAlbumArt())
   eventSource.addEventListener('listeners', (event) => event.data === -1 ? $(LISTENERS_ID).html('N/A') : $(LISTENERS_ID).html(event.data))
   eventSource.addEventListener('listenurl', (event) => {
      if (event.data == '-/-') {
         document.getElementById(STREAM_ID).src = ''
         $(STATUS_ID).html('Disconnected from server.')
      } else {
         streamSrcURL = location.protocol + '//' + event.data
         document.getElementById(STREAM_ID).src = streamSrcURL
         $(STATUS_ID).html(`Connected: <a href='${streamSrcURL}'>${streamSrcURL}</a>`)
      }
   })
}
