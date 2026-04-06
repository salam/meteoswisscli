package api

import "fmt"

// WeatherIconURL returns the MeteoSwiss icon URL for a given icon ID.
// Icon IDs from the plzDetail API correspond to weather condition codes.
func WeatherIconURL(iconID int) string {
	return fmt.Sprintf("https://www.meteoschweiz.admin.ch/static/resources/weather-icons/weather-symbol/%d.svg", iconID)
}

var iconDescriptions = map[int]string{
	1:   "Sunny",
	2:   "Mostly sunny",
	3:   "Partly cloudy",
	4:   "Overcast",
	5:   "Fog",
	6:   "Drizzle",
	7:   "Rain",
	8:   "Snow",
	9:   "Thunderstorm",
	10:  "Sleet",
	11:  "Hail",
	12:  "Light rain",
	13:  "Light snow",
	14:  "Heavy rain",
	15:  "Heavy snow",
	16:  "Rain showers",
	17:  "Snow showers",
	18:  "Thunderstorm with rain",
	19:  "Thunderstorm with snow",
	20:  "Freezing rain",
	21:  "Mixed rain/snow",
	101: "Clear night",
	102: "Mostly clear night",
	103: "Partly cloudy night",
	104: "Overcast night",
	105: "Fog night",
	106: "Drizzle night",
	107: "Rain night",
	108: "Snow night",
	109: "Thunderstorm night",
	126: "Snow showers night",
	132: "Thunderstorm with snow night",
}

func IconDescription(iconID int) string {
	if desc, ok := iconDescriptions[iconID]; ok {
		return desc
	}
	return fmt.Sprintf("Icon %d", iconID)
}
