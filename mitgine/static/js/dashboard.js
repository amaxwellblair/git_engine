$(document).ready(function () {
  retrieveActive();
});

function retrieveActive() {
  $.get("http://localhost:9000/repositories/active", function(data) {
    var repos = JSON.parse(data);
    for (var i = 0; i < repos.length; i++) {
      log(repos[i]);
    }
  });
}

function log( message ) {
  $( "<li class='collection-item'><div>"+message+"</div></li>" ).text( message ).appendTo( ".repo-holder" );
  $( ".repo-holder" ).scrollTop( 0 );
}

function activate(repository) {
  $.post("http://localhost:9000/repositories/activate", { name: repository });
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
