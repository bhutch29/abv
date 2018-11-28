let persist = {};
persist.index = 0;

function changePage(){
    $.getJSON("http://localhost:8081/inventory", function(beers){
        $("#beer-list").empty()
        if (persist.index >= beers.length) {
            persist.index = 0;
        }
        for (var i = persist.index; i < persist.index + 2; i++) {
            if (i == beers.length) {
                break;
            }
            var html = Mustache.to_html($('#beer-entry').html(), beers[i]);
            $('<div class="grid-item"/>').html(html).appendTo('#beer-list')
        }
        persist.index += 2;
    });
};

$(document).ready(changePage);

window.setInterval(function(){
    changePage();
}, 5000);
