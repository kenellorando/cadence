<script>
  import { onMount } from "svelte";

  let title, artist, album, art, search, history, version, listeners, bitrate, listenurl;

  function formatDate(x) {
    var delta = Math.round((+(new Date()) - (new Date(String(x)))) / 1000);
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
    return timeAgo
  }
  let searchRequest = {
    Search: "",
  }
  async function postSearch() {
    try {
      const response = await fetch("http://localhost:8080/api/search", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(searchRequest),
      });

      if (response.ok) {
        console.log("Data posted successfully");
        search = response.message;
      } else {
        console.error("Failed to post data");
      }
    } catch (error) {
      console.error("Error:", error);
    }
  }

  onMount(async () => {
    fetch("http://localhost:8080/api/nowplaying/metadata")
      .then((response) => response.json())
      .then((data) => {
        title = data.Title;
        artist = data.Artist;
        album = data.Album;
      });
    fetch("http://localhost:8080/api/nowplaying/albumart")
      .then((response) => response.json())
      .then((data) => {
        art = data.Picture;
      });
    fetch("http://localhost:8080/api/history")
      .then((response) => response.json())
      .then((data) => {
        history = data;
        alert(history)
      });
    fetch("http://localhost:8080/api/version")
      .then((response) => response.json())
      .then((data) => {
        version = data.Version;
      });
    fetch("http://localhost:8080/api/listeners")
      .then((response) => response.json())
      .then((data) => {
        listeners = data.Listeners;
      });
    fetch("http://localhost:8080/api/bitrate")
      .then((response) => response.json())
      .then((data) => {
        bitrate = data.Bitrate;
      });
    fetch("http://localhost:8080/api/listenurl")
      .then((response) => response.json())
      .then((data) => {
        listenurl = data.ListenURL;
      });
  });
</script>


<!-- Todo: move to player component -->
<div class="font-sans p-4 mx-1 mb-2 rounded-xl shadow-lg text-center border-2 border-neutral-50">
    <div class="p-4">
        <div class=" mb-3 flex justify-center">          
          {#if art == undefined}
            <img class="border-2 border-black w-56" src="/blank.jpg" alt="no album art available">
          {:else}
            <img class="border-2 border-black w-56" src="data:image/jpeg;base64,{art}" alt="album artwork">
          {/if}
        </div>
        <div class="mb-3">
            <div class="text-2xl">{title}</div>
            <div class="text-xl">{artist}</div>
            <div>{album}</div>
        </div>
        <div class="mb-3">
            <button class="bg-gray-200 font-bold py-3 px-5 rounded-full">▶︎</button>
        </div>
        <div class="mb-3">
            <input class="w-50 accent-cyan-600" type="range" name="volume" value="30" min="0" max="100"/>
        </div>
    </div>
</div>


<div class="font-thin font-sans mx-1 mb-2 collapse collapse-arrow rounded-xl border-2 border-neutral-50">
    <input type="checkbox" /> 
    <div class="collapse-title">
      Request
    </div>
    <div class="collapse-content text-sm">
        <div class="form-control w-full max-w-xs">
            <form on:submit|preventDefault={postSearch}>
                <input bind:value={searchRequest.Search} type="text" placeholder="Search for a song or artist!" class="input input-bordered w-full max-w-xs" />
            
            </form>
          </div>
        <div class="overflow-x-auto">
            <table class="table">
              <!-- head -->
              <thead>
                <tr>
                  <th>Title</th>
                  <th>Artist</th>
                  <th>Album</th>
                  <th>Year</th>
                  <th>Request Availability</th>
                </tr>
              </thead>
              <tbody>
                <!-- row 1 -->
                <tr>
                  <td>only my railgun</td>
                  <td>fripSide</td>
                  <td>Infinite Synthesis</td>
                  <td>2009</td>
                  <th>
                    <button class="btn btn-ghost btn-xs">Request</button>
                  </th>
                </tr>
                
              </tbody>
            </table>
            <div class="join">
                <button class="join-item btn">1</button>
                <button class="join-item btn btn-active">2</button>
                <button class="join-item btn">3</button>
                <button class="join-item btn">4</button>
              </div>
          </div>
    </div>
</div>

<div class="font-thin font-sans mx-1 mb-2 collapse collapse-arrow rounded-xl border-2 border-neutral-50">
  <input type="checkbox" /> 
    <div class="collapse-title">
      History
    </div>
    <div class="collapse-content text-sm">
      <div class="overflow-x-auto">
          <table class="table">
            <!-- head -->
            <thead>
              <tr>
                <th>Ended</th>
                <th>Artist</th>
                <th>Title</th>
              </tr>
            </thead>
            <tbody>
              {#if history != undefined}
                {#each history as song}
                <tr>
                  <td>{formatDate(song.Ended)}</td>
                  <td>{song.Artist}</td>
                  <td>{song.Title}</td>
                </tr>
                {/each}
              {:else}
                <p>No track history (yet). The radio may have just started!</p>
              {/if}
            </tbody>
          </table>
    </div>
  </div>
</div>

  <div class="font-thin font-sans mx-1 mb-2 collapse collapse-arrow rounded-xl border-2 border-neutral-50">
    <input type="checkbox" /> 
    <div class="collapse-title">
      UI Theme
    </div>
    <div class="collapse-content text-sm">
        <input type="radio" name="radio-1" class="radio" checked />
        <input type="radio" name="radio-1" class="radio" />
    </div>
  </div>


<div class="font-thin font-sans mx-1 mb-2 collapse collapse-arrow rounded-xl border-2 border-neutral-50">
  <input type="checkbox" /> 
  <div class="collapse-title">
    Radio Information
  </div>
  <div class="collapse-content text-sm font-mono">
      <div>Mountpoint: <span class="link text-cyan-700">{listenurl}</span></div>
      <div>Bitrate (kbps): {bitrate}</div>
      <div>Current Listeners: {listeners}</div>
      <div>Cadence Radio Version: {version}</div>
      <div>
          <a class="link text-cyan-700" target="_blank" href="https://github.com/kenellorando/cadence">GitHub</a> •
          <a class="link text-cyan-700" target="_blank" href="https://github.com/kenellorando/cadence/wiki/API-Reference">API Reference</a>
      </div>
  </div>
</div>
