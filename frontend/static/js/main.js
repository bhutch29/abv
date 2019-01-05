let persist = {};
persist.index = 0;

let defaultTimer = 15000;
let beersPerPage = 16; // must also change CSS grid number

function changePage(){
    $.getJSON("http://" + window.apiUrl + ":8081/inventory/quantity", function(quantity){
        $('#quantity-view').html(quantity + " beers left");
    });

    $.getJSON("http://" + window.apiUrl + ":8081/inventory/variety", function(variety){
        $('#variety-view').html(variety + " varieties to choose from");
    });

    $.getJSON("http://" + window.apiUrl + ":8081/inventory", function(beers){
        $("#beer-list").empty();
        if (persist.index >= beers.length) {
            persist.index = 0;
        }
        for (var i = persist.index; i < persist.index + beersPerPage; i++) {
            if (i == beers.length) {
                break;
            }

            var url = beers[i].Logo;
            var file = url.substring(url.lastIndexOf('/') + 1);
            beers[i].Logo = "/images/" + file;
            var html = Mustache.to_html($('#beer-entry').html(), beers[i]);
            $('<div class="grid-item"/>').html(html).appendTo('#beer-list');

            if (beers[i].Quantity < 3) {
                $(".grid-item").last().attr("id", "quantity-low");
            }
        }

        if (beers.length <= beersPerPage) {
            setTimeout(changePage, 200);
        } else {
            let timer = defaultTimer;
            let mostBeersOnPage = beers.length - persist.index;
            if (mostBeersOnPage < beersPerPage) {
                let ratio = beersPerPage / mostBeersOnPage;
                timer = defaultTimer / ((ratio + 6) / 7);
            }
            persist.index += beersPerPage;
            setTimeout(changePage, timer);
        }

    });
};

function startClock() {
    $('#clock').html(new Date().toLocaleTimeString());
    setTimeout(startClock, 1000);
};

$(document).ready( //registers event last
    $(document).ready(changePage)
);

$(document).ready( //registers event last
    $(document).ready(startClock)
);

