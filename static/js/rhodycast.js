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
    // Setup the chart!
    var chart = new CanvasJS.Chart('waveHeightChart', {
        title:{
            text: "Wave Height (ft)",
            verticalAlign: "top", // "top", "center", "bottom"
            horizontalAlign: "center", // "left", "right", "center"
            fontSize: 15,
            margin: -15,
            fontWeight: "bolder"
        },
        axisY:{
            gridColor: "white",
            gridThickness: 0           
        },
        data: [
        {        
            type: "spline", dataPoints: [
            { x: new Date(2012, 00, 1), y: 1352 },
            { x: new Date(2012, 01, 1), y: 1514 },
            { x: new Date(2012, 02, 1), y: 1321 },
            { x: new Date(2012, 03, 1), y: 1163 },
            { x: new Date(2012, 04, 1), y: 950 },
            { x: new Date(2012, 05, 1), y: 1201 },
            { x: new Date(2012, 06, 1), y: 1186 },
            { x: new Date(2012, 07, 1), y: 1281 },
            { x: new Date(2012, 08, 1), y: 1438 },
            { x: new Date(2012, 09, 1), y: 1305 },
            { x: new Date(2012, 10, 1), y: 1480 },
            { x: new Date(2012, 11, 1), y: 1291 }        
            ]
        }       
        ]
    });

    chart.render();
});