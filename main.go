package main

import (
	"OpenWeatherMap-influx/m/v2/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
)

type Measurements struct {
	Temperature float64 `json:"temp"`
	FeelsLike   float64 `json:"feels_like"`
	TempMin     float64 `json:"temp_min"`
	TempMax     float64 `json:"temp_max"`
	Pressure    uint16  `json:"pressure"`
	Humidity    uint8   `json:"humidity"`
}

type Response struct {
	Coord      json.RawMessage `json:"coord"`      // City geo location
	Weather    json.RawMessage `json:"weather"`    // Weather conditions
	Base       string          `json:"base"`       // Internal parameters
	Main       Measurements    `json:"main"`       // Weather measurements
	Visibility uint16          `json:"visibility"` // Visibility
	Wind       json.RawMessage `json:"wind"`       // Wind measurements
	Clouds     json.RawMessage `json:"clouds"`     // Only contains cloudiness in percentage
	Dt         uint64          `json:"dt"`         // Time of data calculation
	Sys        json.RawMessage `json:"sys"`        // Internal parameters + country code and sunset/sunrise
	Timezone   uint16          `json:"timezone"`   // Shift in seconds from UTC
	ID         uint64          `json:"id"`         // City ID
	Name       string          `json:"name"`       // City name
	Cod        uint16          `json:"cod"`        // Internal parameter
}

func getWeatherData() (temperature float64, humidity uint8, datetime uint64) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", config.APIEndpoint+"weather", nil)

	q := req.URL.Query()
	q.Add("q", config.City+","+config.State)
	q.Add("appid", config.APIToken)
	q.Add("units", "metric")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	// Handle the eventual error
	if err != nil {
		fmt.Println(err)
		return -1, 0, 0
	}

	// Close the response body (Why? Because the docs say so)
	defer resp.Body.Close()

	// If no error is found, get the request body and parse it
	byteValue, _ := ioutil.ReadAll(resp.Body)

	var response Response

	err = json.Unmarshal(byteValue, &response)
	// Handle the eventual error
	if err != nil {
		fmt.Println(err)
		return -1, 0, 0
	}

	return response.Main.Temperature, response.Main.Humidity, response.Dt

}

func sigtermHandler(influx influxdb2.Client) {
	// Prepare to catch the SIGTERM
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Got SIGTERM!")
		// Close InfluxDB Connection
		influx.Close()
		os.Exit(0)
	}()
}

func main() {

	// Create a InfluxDB Client to push the data in the DB
	client := influxdb2.NewClient(config.InfluxHost, config.InfluxToken)
	writeAPI := client.WriteAPI(config.InfluxOrg, config.InfluxBucket)

	// Start the SIGTERM handler
	sigtermHandler(client)

	for {
		// Authenticate and get data
		temperature, humidity, dt := getWeatherData()

		// If the data is invalid, skip them
		if temperature == -1 && humidity == 0 {
			time.Sleep(config.RequestDelay * time.Second)
			continue
		}

		// Get the current time (to add to the data)
		relTime := time.Unix(int64(dt), 0)

		// Create the point with all the data
		p := influxdb2.NewPointWithMeasurement(config.InfluxMeasurementName).
			AddTag("serviceName", "OpenWeatherMap").
			AddField("temperature", temperature).
			AddField("humidity", int16(humidity)).
			SetTime(relTime)
		writeAPI.WritePoint(p)
		writeAPI.Flush()

		time.Sleep(config.RequestDelay * time.Second)
	}
}
