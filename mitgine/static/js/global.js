$(document).ready(function() {
  displayNav();
});

function displayNav() {
  var url = document.URL;
  var path = url.split("/").pop();
  if (path != "") {
    $("#logout").show();
    $("#logo").show();
  }
}
