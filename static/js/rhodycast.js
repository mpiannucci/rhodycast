$(document).ready(function() {
    // Setup the chart!
    $.ajax({
        url: '/forecast_as_json',
        type: 'GET'
    }).done(function(data) {

        stripLines = [];
        dataSeries = [];
        for (var i = 0; i < data.ForecastData.length; i++) {
            var waveHeightRange = roundHundreths(data.ForecastData[i].MinimumBreakingHeight) + " - " + roundHundreths(data.ForecastData[i].MaximumBreakingHeight) + " ft";
            var swell = roundHundreths(data.ForecastData[i].PrimarySwellComponent.WaveHeight) + " @ " + roundHundreths(data.ForecastData[i].PrimarySwellComponent.Period) + " s " + roundHundreths(data.ForecastData[i].PrimarySwellComponent.Direction) + "\xB0";
            var wind  = roundHundreths(data.ForecastData[i].WindDirection) + "\xB0 " + roundHundreths(data.ForecastData[i].WindSpeed) + " mph";
            var dateString = data.ForecastData[i].Date.split(" ")[0] + " " + data.ForecastData[i].Time;

            xLabel = " ";

            if (stripLines.length === 0) {
                if ((data.ForecastData[i].Time === "07 PM") || (data.ForecastData[i].Time === "08 PM") ||
                    (data.ForecastData[i].Time === "01 AM") || (data.ForecastData[i].Time === "02 AM")) {
                    stripLineNegative = {
                        startValue: i,
                        endValue: 0,
                        color: "#F2F2F2",
                    }
                    stripLines.push(stripLineNegative);
                } else {
                    var day = data.ForecastData[i].Date.split(" ")[0];
                    var labelSize = 12
                    var width = (window.innerWidth > 0) ? window.innerWidth : screen.width;
                    if ((width < 600) && (width > 400)) {
                        day = day.substring(0,3);
                    } else if (width <= 400) {
                        day = day.substring(0,3);
                    labelSize = 10;
                    }

                    stripLine = {
                        startValue: i,
                        endValue: 0,
                        color: "#FFFFFF",
                        label: day,
                        labelBackgroundColor: "#FFFFFF",
                        labelFontColor: "#838383",
                        labelFontSize: labelSize
                    }
                }
            } else {
                if ((data.ForecastData[i].Time === "07 PM") || (data.ForecastData[i].Time === "08 PM")) {
                    stripLines[stripLines.length - 1].endValue = i;
                    stripLineNegative = {
                        startValue: i,
                        endValue: 0,
                        color: "#F2F2F2",
                    }
                    stripLines.push(stripLineNegative);
                } else if ((data.ForecastData[i].Time === "04 AM") || (data.ForecastData[i].Time === "05 AM")) {
                    stripLines[stripLines.length - 1].endValue = i;

                    var day = data.ForecastData[i].Date.split(" ")[0];
                    var labelSize = 12
                    var width = (window.innerWidth > 0) ? window.innerWidth : screen.width;
                    if ((width < 600) && (width > 400)) {
                        day = day.substring(0,3);
                    } else if (width <= 400) {
                        day = day.substring(0,3);
                        labelSize = 10;
                    }
                    stripLine = {
                        startValue: i,
                        endValue: 0,
                        color: "#FFFFFF",
                        label: day,
                        labelBackgroundColor: "#FFFFFF",
                        labelFontColor: "#838383",
                        labelFontSize: labelSize
                    }
                    stripLines.push(stripLine);
                }
            }

            dataPoint = {
                y: Math.round(data.ForecastData[i].MaximumBreakingHeight * 100) / 100,
                x: i,
                label: xLabel,
                toolTipContent: "<b>" + dateString + ": " + waveHeightRange + "</b><br>Swell: " + swell + "<br>Wind: " + wind
            }
            dataSeries.push(dataPoint);
        }

        stripLines[stripLines.length - 1].endValue = data.ForecastData.length;

        var chart = new CanvasJS.Chart('waveHeightChart', {
            title: {
                text: "Wave Height (ft)",
                verticalAlign: "top", 
                horizontalAlign: "center",
                fontSize: 15,
                margin: 3,
                fontWeight: "bolder"
            },
            axisY: {
                gridColor: "#F2F2F2",
                lineColor: "#F2F2F2",
                tickColor: "#F2F2F2",
                gridThickness: 1,
                labelFontSize: 20         
            },
            axisX: {
                gridColor: "#F2F2F2",
                lineColor: "#F2F2F2",
                tickColor: "#F2F2F2",
                gridThickness: 0,
                interval: 1,
                stripLines: stripLines
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

function roundHundreths(original) {
    return Math.round(original * 100.0) / 100.0;
}