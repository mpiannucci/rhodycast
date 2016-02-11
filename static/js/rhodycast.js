function initMap() {
    var latitude = parseFloat($('#map').attr('data-lat'));
    var longitude = parseFloat($('#map').attr('data-lng'));
    
    map = new google.maps.Map($('#map')[0], {
        center: {lat: latitude, lng: longitude},
        zoom: 10,
        mapTypeId: google.maps.MapTypeId.HYBRID
    });

    var spotMarker = new google.maps.Marker({
        position: {lat: latitude, lng: longitude},
        map: map,
        title: $('#map').attr('data-name')
    });
}

$(document).ready(function() {
    // Setup stuff!
});