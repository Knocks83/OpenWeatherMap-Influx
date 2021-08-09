package config

// OpenWeatherMap
const (
	APIToken = ""
	City     = ""
	State    = ""
)

// InfluxDB
const (
	InfluxHost            = "http://localhost:8086" // The host where the Influx DB is (default port: 8086)
	InfluxToken           = ""                      // The access token (if using 1.8>=version<2.0 use user:pass as token)
	InfluxOrg             = ""                      // The organization (if using 1.8>=version<2.0 leave empty)
	InfluxBucket          = ""                      // The bucket (if using 1.8>=version<2.0 use database/retention-policy, or just the db if the default rp should be used)
	InfluxMeasurementName = ""                      // The name of the Influx Measurement
)

// DO NOT EDIT ANYTHING UNDER THIS COMMENT IF YOU DON'T KNOW WHAT YOU'RE DOING

// OpenWeatherMap
const (
	APIEndpoint  = "https://api.openweathermap.org/data/2.5/" // Endpoint used to send requests to the Thermostat
	RequestDelay = 600                                        // Make a request every 10 minutes (that's the update rate on their platform)
)
