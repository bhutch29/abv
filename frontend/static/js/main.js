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

            setImagePath(beers[i]);
            createBeerEntry(beers[i]);
            setLowQuantityIndication(beers[i]);
        }

        setPageNumber(beers.length);
        setPageTimeout(beers.length);
        persist.index += beersPerPage;
    });
};

function setImagePath(beer) {
    var url = beer.Logo;
    var file = url.substring(url.lastIndexOf('/') + 1);
    beer.Logo = "/images/" + file;
}

function createBeerEntry(beer) {
    var html = Mustache.to_html($('#beer-entry').html(), beer);
    $('<div class="grid-item"/>').html(html).appendTo('#beer-list');
}

function setLowQuantityIndication(beer) {
    if (beer.Quantity < 3) {
        $(".grid-item").last().attr("id", "quantity-low");
    }
}

function setPageNumber(numBeers) {
    var numPages = Math.ceil(numBeers/beersPerPage);
    var currentPage = Math.ceil((persist.index + 1) / beersPerPage);
    $('#page-number-view').html("Page " + currentPage + "/" + numPages);
}

function setPageTimeout(numBeers) {
        if (numBeers <= beersPerPage) { //Only 1 page
            setTimeout(changePage, 200);
        } else {
            timer = calcPageTimer(numBeers);
            setTimeout(changePage, timer);
        }
}

function calcPageTimer(numBeers) {
    let timer = defaultTimer;
    let mostBeersOnPage = numBeers - persist.index;
    if (mostBeersOnPage < beersPerPage) {
        let ratio = beersPerPage / mostBeersOnPage;
        timer = defaultTimer / ((ratio + 6) / 7);
    }
    return timer;
}

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

