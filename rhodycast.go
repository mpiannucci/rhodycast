package rhodycast

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"

	"github.com/mpiannucci/surfnerd"
)

var funcMap = template.FuncMap{
	"DegreeToDirection":  surfnerd.DegreeToDirection,
	"ToTwelveHourFormat": surfnerd.ToTwelveHourFormat,
	"ToFixedPoint":       ToFixedPoint,
}

var indexTemplate = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/index.html"))

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/___fetch___", waveWatchFetchHandler)
	http.HandleFunc("/forecast_as_json", forecastJsonHandler)
	http.HandleFunc("/modeldata_as_json", modelDataJsonHandler)
}

// forecastKey returns the key used for all forecast entries.
func forecastKey(c context.Context) *datastore.Key {
	// The string "default_forecast" here could be varied to have multiple forecasts.
	return datastore.NewKey(c, "Forecast", "default_forecast", 0, nil)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixedPoint(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Query the current count of forecasts
	q := datastore.NewQuery("Forecast")
	var forecasts []surfnerd.WaveWatchForecast
	_, keyError := q.GetAll(ctx, &forecasts)
	if keyError != nil {
		http.Error(w, keyError.Error(), http.StatusInternalServerError)
		return
	}

	if len(forecasts) < 1 {
		http.Error(w, "No forecasts available", http.StatusInternalServerError)
		return
	}

	if err := indexTemplate.Execute(w, forecasts[0]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func waveWatchFetchHandler(w http.ResponseWriter, r *http.Request) {
	ctxParent := appengine.NewContext(r)
	ctx, _ := context.WithTimeout(ctxParent, 60*time.Second)
	client := urlfetch.Client(ctx)

	// Set the location to fetch from
	riLocation := surfnerd.Location{
		Latitude:     40.969,
		Longitude:    360 - 71.127,
		Elevation:    0,
		LocationName: "Block Island Buoy - 44097",
	}

	// Create the model and get its url to fetch the latest data from
	wwModel := surfnerd.GetWaveModelForLocation(riLocation)
	wwURL := wwModel.CreateURL(riLocation, 0, 60)

	resp, httpErr := client.Get(wwURL)
	if httpErr != nil {
		http.Error(w, httpErr.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "HTTP GET returned status %v\n", resp.Status)
	defer resp.Body.Close()

	// Read all of the raw data
	contents, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		http.Error(w, readErr.Error(), http.StatusInternalServerError)
		return
	}

	// Put the forecast data into containers
	modelData := surfnerd.WaveWaveModelDataFromRaw(riLocation, contents)
	forecast := surfnerd.WaveWatchForecastFromModelData(modelData)
	if forecast == nil {
		http.Error(w, "Error parsing wavewatch data", http.StatusInternalServerError)
		return
	}

	// Convert to imperial
	forecast.ConvertToImperialUnits()

	// Query the current count of forecasts
	q := datastore.NewQuery("Forecast")
	entryCount, countError := q.Count(ctxParent)
	if countError != nil {
		http.Error(w, countError.Error(), http.StatusInternalServerError)
		return
	}

	// If there is an entity then swap it out with the new one. Otherwise make a
	// new one
	if entryCount > 0 {
		var forecasts []surfnerd.WaveWatchForecast
		keys, keyError := q.GetAll(ctxParent, &forecasts)
		if keyError != nil {
			http.Error(w, keyError.Error(), http.StatusInternalServerError)
		}
		datastore.Put(ctxParent, keys[0], forecast)
	} else {
		// Get the datastore key from the default forecast entry
		key := datastore.NewIncompleteKey(ctxParent, "Forecast", forecastKey(ctxParent))
		if _, putErr := datastore.Put(ctxParent, key, forecast); putErr != nil {
			http.Error(w, putErr.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func forecastJsonHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Query the current count of forecasts
	q := datastore.NewQuery("Forecast")
	var forecasts []surfnerd.WaveWatchForecast
	_, keyError := q.GetAll(ctx, &forecasts)
	if keyError != nil {
		http.Error(w, keyError.Error(), http.StatusInternalServerError)
		return
	}

	if len(forecasts) < 1 {
		http.Error(w, "No forecasts available", http.StatusInternalServerError)
		return
	}

	forecast := forecasts[0]
	forecastJson, jsonErr := forecast.ToJSON()
	if jsonErr != nil {
		http.Error(w, "Could not marshal Forecast to json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(forecastJson)
}

func modelDataJsonHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Query the current count of forecasts
	q := datastore.NewQuery("Forecast")
	var forecasts []surfnerd.WaveWatchForecast
	_, keyError := q.GetAll(ctx, &forecasts)
	if keyError != nil {
		http.Error(w, keyError.Error(), http.StatusInternalServerError)
		return
	}

	if len(forecasts) < 1 {
		http.Error(w, "No forecasts available", http.StatusInternalServerError)
		return
	}

	forecast := forecasts[0]
	modelData := forecast.ToModelData()
	modelDataJson, jsonErr := modelData.ToJSON()
	if jsonErr != nil {
		http.Error(w, "Could not marshal ModelData to json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(modelDataJson)
}
