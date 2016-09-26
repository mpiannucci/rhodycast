package rhodycast

import (
	"errors"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/mpiannucci/surfnerd"
)

var funcMap = template.FuncMap{
	"DegreeToDirection":  surfnerd.DegreeToDirection,
	"ToTwelveHourFormat": surfnerd.ToTwelveHourFormat,
	"ToFixedPoint":       ToFixedPoint,
}

var indexTemplate = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/index.html"))
var aboutTemplate = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/about.html"))

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/___fetch___", modelFetchHandler)
	http.HandleFunc("/forecast_as_json", forecastJsonHandler)
}

// forecastKey returns the key used for all forecast entries.
func forecastKey(c context.Context) *datastore.Key {
	// The string "default_forecast" here could be varied to have multiple forecasts.
	return datastore.NewKey(c, "SurfForecast", "default_forecast", 0, nil)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixedPoint(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func fetchWaveForecast(loc surfnerd.Location, client *http.Client, w http.ResponseWriter) (*surfnerd.WaveForecast, error) {
	waveModel := surfnerd.GetWaveModelForLocation(loc)
	waveURL := waveModel.CreateURL(loc, 0, 60)

	// Fetch the wave data!
	waveResp, waveHttpErr := client.Get(waveURL)
	if waveHttpErr != nil {
		http.Error(w, waveHttpErr.Error(), http.StatusInternalServerError)
		return nil, waveHttpErr
	}
	defer waveResp.Body.Close()

	// Read all of the raw data
	waveContents, waveReadErr := ioutil.ReadAll(waveResp.Body)
	if waveReadErr != nil {
		http.Error(w, waveReadErr.Error(), http.StatusInternalServerError)
		return nil, waveReadErr
	}

	// Put the forecast data into containers
	waveModelData := surfnerd.WaveModelDataFromRaw(loc, waveModel.NOAAModel, waveContents)
	waveForecast := surfnerd.WaveForecastFromModelData(waveModelData)
	if waveForecast == nil {
		http.Error(w, "Error parsing wavewatch data", http.StatusInternalServerError)
		return nil, errors.New("Wave Forecast is Nil")
	}

	return waveForecast, nil
}

func fetchWindForecast(loc surfnerd.Location, client *http.Client, w http.ResponseWriter) (*surfnerd.WindForecast, error) {
	windModel := surfnerd.GetWindModelForLocation(loc)
	windURL := windModel.CreateURL(loc, 0, 60)

	// Fetch the wind data
	windResp, windHttpErr := client.Get(windURL)
	if windHttpErr != nil {
		http.Error(w, windHttpErr.Error(), http.StatusInternalServerError)
		return nil, windHttpErr
	}
	defer windResp.Body.Close()

	// Read all of the raw data
	windContents, windReadErr := ioutil.ReadAll(windResp.Body)
	if windReadErr != nil {
		http.Error(w, windReadErr.Error(), http.StatusInternalServerError)
		return nil, windReadErr
	}

	// Put the forecast data into containers
	windModelData := surfnerd.WindModelDataFromRaw(loc, windModel.NOAAModel, windContents)
	windForecast := surfnerd.WindForecastFromModelData(windModelData)
	if windForecast == nil {
		http.Error(w, "Error parsing wavewatch data", http.StatusInternalServerError)
		return nil, errors.New("Wind Forecast is Nil")
	}

	return windForecast, nil
}

func modelFetchHandler(w http.ResponseWriter, r *http.Request) {
	ctxParent := appengine.NewContext(r)
	ctx, _ := context.WithTimeout(ctxParent, 480*time.Second)
	client := urlfetch.Client(ctx)

	// Set the location to fetch from
	riWaveLocation := surfnerd.Location{
		Latitude:     41.323,
		Longitude:    360 - 71.396,
		Elevation:    30.0,
		LocationName: "Block Island Sound",
	}
	riWindLocation := surfnerd.Location{
		Latitude:     41.6,
		Longitude:    360 - 71.500,
		Elevation:    10,
		LocationName: "Narragansett Pier",
	}
	riForecastLocation := surfnerd.Location{
		Latitude:     41.395,
		Longitude:    -71.453,
		LocationName: "Narragansett",
	}

	waveForecast, waveError := fetchWaveForecast(riWaveLocation, client, w)
	windForecast, windError := fetchWindForecast(riWindLocation, client, w)

	if waveError != nil {
		log.Errorf(ctxParent, waveError.Error())
		http.Error(w, waveError.Error(), http.StatusInternalServerError)
		return
	}

	if windError != nil {
		log.Errorf(ctxParent, windError.Error())
		http.Error(w, windError.Error(), http.StatusInternalServerError)
		return
	}

	surfForecast := surfnerd.NewSurfForecast(riForecastLocation, 145.0, 0.02, waveForecast, windForecast)
	surfForecast.ChangeUnits(surfnerd.English)

	// Query the current count of forecasts
	q := datastore.NewQuery("SurfForecast")
	entryCount, countError := q.Count(ctxParent)
	if countError != nil {
		http.Error(w, countError.Error(), http.StatusInternalServerError)
		return
	}

	// If there is an entity then swap it out with the new one. Otherwise make a
	// new one
	if entryCount > 0 {
		var forecasts []surfnerd.SurfForecast
		keys, keyError := q.GetAll(ctxParent, &forecasts)
		if keyError != nil {
			http.Error(w, keyError.Error(), http.StatusInternalServerError)
		}
		datastore.Put(ctxParent, keys[0], surfForecast)
	} else {
		// Get the datastore key from the default forecast entry
		key := datastore.NewIncompleteKey(ctxParent, "SurfForecast", forecastKey(ctxParent))
		if _, putErr := datastore.Put(ctxParent, key, surfForecast); putErr != nil {
			http.Error(w, putErr.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Query the current count of forecasts
	q := datastore.NewQuery("SurfForecast")
	var forecasts []surfnerd.SurfForecast
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

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	if err := aboutTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func forecastJsonHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Query the current count of forecasts
	q := datastore.NewQuery("SurfForecast")
	var forecasts []surfnerd.SurfForecast
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
