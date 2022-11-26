function getSoon(site) {
  var launchTime = new Date(2025, 0, 1, 23, 59); // Set launch: [year], [month], [day], [hour]...
  var raw = new XMLHttpRequest()
  raw.open('GET', 'https://cms.chewedfeed.com/service/' + site, false)
  raw.onreadystatechange = function() {
    if (raw.readyState === 4) {
      if (raw.status === 200 || raw.status !== 0) {
        var allText = raw.responseText;
        var json = JSON.parse(allText);
        var html = '';
        var year = json.launchDate.year
        var month = json.launchDate.month
        var day = json.launchDate.day
        var newLaunch = new Date(year, month, day, 23, 59)

        if (newLaunch.getTime() < launchTime.getTime()) {
          launchTime = newLaunch
        }

        $("#countdown").countdown({
          until: launchTime,
          format: "odHMS"
        });
        $("#progressbar").animate({width: json.progress + "%"}, 2000);
        $("#progressAmount").text(json.progress + "%");
        $("#pageTitle").text(json.name + " Coming Soon");

        if (showdown !== undefined) {
          var s = new showdown.Converter();
          if (s !== undefined) {
            var html = s.makeHtml(json.fullDesc);
            var el = document.getElementById("fullDesc");
            el.innerHTML = html;
          }
        }
      }
    }
  }
  raw.send(null);
}
