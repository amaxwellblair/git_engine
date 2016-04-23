$(document).ready(function () {

});

$(function() {
  function log( message ) {
    $( "<li class='collection-item'><div>"+message+"</div></li>" ).text( message ).appendTo( ".repo-holder" );
    $( ".repo-holder" ).scrollTop( 0 );
  }
  function activate(repository) {
    $.post("http://localhost:9000/repositories/activate", { name: repository }, function(data) {
      console.log(data);
    });
  }

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
