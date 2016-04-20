$(document).ready(function () {
  $.getJSON("http://localhost:9000/repositories", function (data) {
    console.log(data)
  });
});
