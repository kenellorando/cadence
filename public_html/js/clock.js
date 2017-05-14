function checkTime(i) {
  if (i < 10) {
    i = "0" + i;
  } // add zero in front of numbers < 10
  return i;
}

function clock() {
  var time = new Date();
  var currentHour = time.getHours();
  var currentMinute = time.getMinutes();
  var currentSecond = time.getSeconds();
  //var currentMillisecond = time.getMilliseconds();
  
  currentMinute = checkTime(currentMinute);
  currentSecond = checkTime(currentSecond);
  
  
  document.getElementById("clock").innerHTML = currentHour + ":" + currentMinute + ":" + currentSecond;
}

window.setInterval(function () {
  clock();
  serverClock();
}, 1000);
