$(document).ready(function () {
  retrieveActive();
  load_repos();
});

$(".refresh-button").click(function() {
  $.get("http://localhost:9000/refresh/repositories")
})

function log( message ) {
  var url = ("http://localhost:9000/dashboard/" + message);
  $( "<a class='collection-item' href='"+url+"'>"+message+"</a>" ).text( message ).appendTo( ".repo-holder" );
  $( ".repo-holder" ).scrollTop( 0 );
}

function retrieveActive() {
  $.get("http://localhost:9000/repositories/active", function(data) {
    var repos = JSON.parse(data);
    for (var i = 0; i < repos.length; i++) {
      log(repos[i]);
    }
  });
}

function activate(repository) {
  $.post("http://localhost:9000/repositories/activate", { name: repository });
}

function load_repos() {
  $.get("http://localhost:9000/repositories")
}

$(function() {
  $( "#repository" ).autocomplete({
    source: "http://localhost:9000/repositories",
    minLength: 2,
    select: function( event, ui ) {
      log( ui.item ?
        ui.item.value :
        "Nothing selected, input was " + this.value );
      activate(ui.item.value);
    }
  });
});
