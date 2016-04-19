$(document).ready(function() {
  displayLogout();
});

function displayLogout() {
  var url = document.URL;
  var path = url.split("/").pop();
  if (path != "") {
    $("#logout").show();
  }
}
