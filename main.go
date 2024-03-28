package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var apiKey string

var zeroC = 273.15

type WeatherData struct {
	Weather []Weather
	Main    DataPoints
	Name    string
}

type Weather struct {
	Main        string
	Description string
}

type DataPoints struct {
	Temp float64
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGet(w, r)
	} else if r.Method == http.MethodPost {
		handlePost(w, r)
	} else {
		http.Error(w, "This endpoint only support GET and POST requests", http.StatusMethodNotAllowed)
		return
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	lat := matchParam(params, acceptableLats)
	long := matchParam(params, acceptableLongs)
	if len(lat) != 1 || len(long) != 1 {
		http.Error(w, "Valid coordinates not provided. Please provide exactly one floating-point number for each.", http.StatusBadRequest)
		return
	}

	log.Info().Msg("Fetching weather data...")
	log.Debug().Str("lat", lat[0]).Str("long", long[0]).Msg("User-provided coordinates from query params")

	res, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&appid=%s", lat[0], long[0], apiKey))
	if err != nil || res.StatusCode >= 400 {
		http.Error(w, "Error fetching weather data.", http.StatusBadGateway)
		return
	}

	log.Debug().Str("status code", fmt.Sprintf("%d", res.StatusCode)).Msg("response status code")

	body, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		http.Error(w, "Error parsing JSON response", http.StatusInternalServerError)
		return
	}

	log.Debug().Msg(reportWeather(weatherData))
	log.Debug().Str("temp", fmt.Sprintf("%f", weatherData.Main.Temp-zeroC)).Msg("Temperature in Celsius")

	_, err = io.WriteString(w, reportWeather(weatherData))
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	}
}

func reportWeather(wd WeatherData) string {
	rawTemp := wd.Main.Temp - zeroC
	var temperature string
	if rawTemp < 10 {
		temperature = "cold"
	} else if rawTemp >= 10 && rawTemp < 20 {
		temperature = "moderate"
	} else {
		temperature = "hot"
	}

	weather := ""
	for idx, w := range wd.Weather {
		if idx == len(wd.Weather)-1 && len(wd.Weather) != 1 {
			weather += "and "
		}

		weather += w.Description + ", "
	}

	return fmt.Sprintf("The weather in %s is %sand the temperature is %s.", wd.Name, weather, temperature)
}

var acceptableLats = []string{
	"lat",
	"LAT",
	"lattitude",
	"Lattitude",
	"LAttitude",
	"LATTITUDE",
}

var acceptableLongs = []string{
	"lon",
	"long",
	"longitude",
	"Longitude",
	"LOngitude",
	"LONGITUDE",
}

func matchParam(params map[string][]string, matches []string) []string {
	for _, s := range matches {
		if val, ok := params[s]; ok {
			return val
		}
	}
	return []string{}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	//TODO: Accept coordinates POSTed with JSON
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msg("Starting up weather fetcher...")

	apiKey = os.Getenv("OPENWEATHER_API_KEY")

	if apiKey == "" {
		fmt.Println("You must set the OPENWEATHER_API_KEY environment variable")
		os.Exit(1)
	}

	servePort := os.Getenv("OW_HTTP_PORT")
	if servePort == "" {
		servePort = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("/weather", http.HandlerFunc(handle))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", servePort), mux); err != nil {
		log.Fatal().Err(err)
	}
}
