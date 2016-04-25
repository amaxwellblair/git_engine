$('#search').keypress(function (e) {
  if (e.which == 13) {
    var input = $('#search').val();
    get_commits(input);
    return false;
  }
});

function get_commits(search) {
  $.get(commits_url(), function (data) {
    var commits = JSON.parse(data);
    put_commits(commits);
  });
}

function log(commit) {
  var url = (commit["html_url"]);
  $("<a class='collection-item' href='"+url+"'>"+commit["message"]+"</a>").text(commit["message"]).appendTo(".commit-holder");
  $(".commit-holder").scrollTop(0);
}

function put_commits(commits) {
  for (var i = 0; i < commits.length; i++) {
    log(commits[i]);
  }
}

function commits_url() {
  var bits = document.URL.split("/");
  var repo = bits[bits.length - 1];
  return "http://localhost:9000/"+repo+"/commits";
}
