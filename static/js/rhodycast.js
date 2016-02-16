$(document).ready(function() {
    // Setup the chart!
    $.ajax({
        url: '/forecast_as_json',
        type: 'GET'
    }).done(function(data) {

        // Create some ranges so we can make the chart consistently colored
        var firstTime = data.ForecastData[0].Time;
        var splitStartIndex = 0;
        var splitLength = 4;
        var firstSplitLength = splitLength;
        if ((firstTime === "01 AM") || (firstTime === "02 AM")){
            // Defaults are good
            splitStartIndex = 0;
        } else if (firstTime === "12z") {
            firstSplitLength = 2;
        } else if (firstTime === "18z") {
            splitStartIndex = 4;
        } else if (firstTime === "00z") {
            splitStartIndex = 2;
        }

        var stripLines = []
        var nextStartValue = 0;
        var firstFlag = true;
        var dayIndex = 0;
        while (nextStartValue < data.ForecastData.length) {
            var day = " ";
            if (firstFlag) {
                day = data.ForecastData[splitStartIndex].Date.split(" ")[0];
            } else {
                day = data.ForecastData[nextStartValue].Date.split(" ")[0];
            }

            var labelSize = 12
            var width = (window.innerWidth > 0) ? window.innerWidth : screen.width;
            if ((width < 600) && (width > 400)) {
                day = day.substring(0,3);
            } else if (width <= 400) {
                day = day.substring(0,3);
                labelSize = 10;
            }

            stripLine = {
                startValue: nextStartValue,
                endValue: nextStartValue + splitLength,
                color: "#FFFFFF",
                label: day,
                labelBackgroundColor: "#FFFFFF",
                labelFontColor: "#838383",
                labelFontSize: labelSize
            }

            if (firstFlag && (splitStartIndex == 0)) {
                stripLine.startValue = splitStartIndex;
                stripLine.endValue = splitStartIndex + firstSplitLength;
                nextStartValue += firstSplitLength;
                firstFlag = false;
            } else {
                nextStartValue += splitLength
            }

            stripLines.push(stripLine);

            stripLineNegative = {
                startValue: nextStartValue,
                endValue: nextStartValue + splitLength,
                color: "#F2F2F2",
            }

            if (firstFlag) {
                stripLineNegative.startValue = 0;
                stripLineNegative.endValue = splitStartIndex;
                nextStartValue += splitStartIndex;
                firstFlag = false;
            } else {
                nextStartValue += splitLength;
            }

            stripLines.push(stripLineNegative);
        }

        dataSeries = [];
        for (var i = 0; i < data.ForecastData.length; i++) {
            var waveHeightRange = roundHundreths(data.ForecastData[i].MinimumBreakingHeight) + " - " + roundHundreths(data.ForecastData[i].MaximumBreakingHeight) + " ft";
            var swell = roundHundreths(data.ForecastData[i].PrimarySwellComponent.WaveHeight) + " @ " + roundHundreths(data.ForecastData[i].PrimarySwellComponent.Period) + " s " + roundHundreths(data.ForecastData[i].PrimarySwellComponent.Direction) + "\xB0";
            var wind  = roundHundreths(data.ForecastData[i].WindDirection) + "\xB0 " + roundHundreths(data.ForecastData[i].WindSpeed) + " mph";
            var dateString = data.ForecastData[i].Date.split(" ")[0] + " " + data.ForecastData[i].Time;

            xLabel = " ";

            dataPoint = {
                y: Math.round(data.ForecastData[i].MaximumBreakingHeight * 100) / 100,
                x: i,
                label: xLabel,
                toolTipContent: "<b>" + dateString + ": " + waveHeightRange + "</b><br>Swell: " + swell + "<br>Wind: " + wind
            }
            dataSeries.push(dataPoint);
        }

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
                interval: 4,
                stripLines: []
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
