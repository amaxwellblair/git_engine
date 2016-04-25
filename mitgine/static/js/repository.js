$('#search').keypress(function (e) {
  if (e.which == 13) {
    var input = $('#search').val();
    clear_log();
    get_commits(input);
    return false;
  }
});

function get_commits(search) {
  $.get(commits_url(search), function (data) {
    var commits = JSON.parse(data);
    put_commits(commits);
  });
}

function log(commit) {
  var url = commit["html_url"];
  var message = commit["commit_message"];
  $("<a class='collection-item' target='_blank' href='"+url+"'>"+message+"</a>").text(message).appendTo(".commit-holder");
  $(".commit-holder").scrollTop(0);
}

function no_log() {
  var message = "No commits found...";
  $("<li class='collection-item'>"+message+"</li>").text(message).appendTo(".commit-holder");
  $(".commit-holder").scrollTop(0);
}

function clear_log() {
  $(".commit-holder").empty();
}

function put_commits(commits) {
  if (commits != null) {
    for (var i = 0; i < commits.length; i++) {
      log(commits[i]);
    }
  } else {
      no_log();
  }
}

function commits_url(term) {
  var bits = document.URL.split("/");
  var repo = bits[bits.length - 1];
  return "http://localhost:9000/dashboard/"+repo+"/commits?term="+term;
}
