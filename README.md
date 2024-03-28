# Simple Weather Reporter

This tool is a web server that fetches data from [OpenWeather](https://openweathermap.org/) and reports on the weather accordingly. Send it a request with the lattitude & longitude of a particular place, and you will receive a simple weather report in return.

### Requirements
* Go >=1.21

To use, clone this repo, set the env var `OPENWEATHER_API_KEY` to your API key and run the code:

```
$ OPENWEATHER_API_KEY=<key> go run main.go
```

This will lunch a web server that binds to port 8080 by default. This can be overridden with the env var `OW_HTTP_PORT`.

Requests must be directed to `/weather`, and query params for `lat` and `long` must be provided for lattitude and longitude, respectively, e.g.:

```
$ curl localhost:8080/weather\?lat=40.7128\&long=-74.0060
The weather in New York is mist, and the temperature is cold.
```
