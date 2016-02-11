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
    $.ajax({
        url: '/forecast_as_json',
        type: 'GET'
    }).done(function(data) {

        dataSeries = [];
        for (var i = 0; i < data.ForecastData.length; i++) {
            var waveHeightRange = Math.round(data.ForecastData[i].MinimumBreakingHeight * 100) / 100 + " - " + Math.round(data.ForecastData[i].MaximumBreakingHeight * 100) / 100 + " ft";
            var swell = Math.round(data.ForecastData[i].SignificantWaveHeight * 100) / 100 + " @ " + Math.round(data.ForecastData[i].MeanWavePeriod * 100) / 100 + " s " + Math.round(data.ForecastData[i].DominantWaveDirection * 100) / 100 + "\xB0";
            var wind  = Math.round(data.ForecastData[i].SurfaceWindDirection * 100) / 100 + "\xB0 " + Math.round(data.ForecastData[i].SurfaceWindSpeed * 100) / 100 + " mph";

            dataPoint = {
                y: Math.round(data.ForecastData[i].MaximumBreakingHeight * 100) / 100,
                x: i,
                label: data.ForecastData[i].Date.split(" ")[0] + " " + data.ForecastData[i].Time,
                toolTipContent: "<b>{label}: " + waveHeightRange + "</b><br>Swell: " + swell + "<br>Wind: " + wind
            }
            dataSeries.push(dataPoint);
        }

        var chart = new CanvasJS.Chart('waveHeightChart', {
            title: {
                text: "Wave Height (ft)",
                verticalAlign: "top", 
                horizontalAlign: "center",
                fontSize: 15,
                margin: -15,
                fontWeight: "bolder"
            },
            axisY: {
                gridColor: "white",
                lineColor: "#F2F2F2",
                tickColor: "#F2F2F2",
                gridThickness: 0           
            },
            axisX: {
                lineColor: "#F2F2F2",
                tickColor: "#F2F2F2",
                labelAutoFit: true,
                labelFontSize: 11
            },
            data: [
            {        
                type: "spline", 
                dataPoints: dataSeries,
                color: "#0D9EFF"
            }       
            ], 
        });

        chart.render();
    });
});