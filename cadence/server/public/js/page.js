$(document).ready(() => {
   const PLAY_BUTTON = 'playButton'
   const VOLUME_KEY = 'volumeKey'

   const stream = document.getElementById('stream')
   const volbar = document.getElementById('volume')

   // Warn iOS and Safari users
   let safariUA = /Apple/i.test(navigator.vendor)
   let iOSUA = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream

   iOSUA || safariUA && alert('You appear to be using an iOS device or a Safari browser. Cadence stream playback may not be compatible with your platform.')

   let play = (stream) => {
      stream.src = streamSrcURL
      stream.load()
      stream.play()
      document.getElementById(PLAY_BUTTON).innerHTML = '⏸'
   }

   let pause = (stream) => {
      stream.src = ''
      stream.load()
      stream.pause()
      document.getElementById(PLAY_BUTTON).innerHTML = '⏵'
   }

   // Loads and unloads audio stream source
   document
      .getElementById(PLAY_BUTTON)
      .addEventListener('click', () => stream.paused ? play() : pause())

   // Load cached volume level, or 30%
   volbar.value = stream.volume = localStorage.getItem(VOLUME_KEY) || 0.3

   // Volume bar listeners
   $('#volume')
      .change(() => {
         stream.volume = this.value
         localStorage.setItem(VOLUME_KEY, this.value)
      })
      .on('input', () => {
         stream.volume = this.value
         localStorage.setItem(VOLUME_KEY, this.value)
      })

   // Search keyup
   $('#searchInput').keyup((event) => event.keyCode == 13 && postSearch())
})
