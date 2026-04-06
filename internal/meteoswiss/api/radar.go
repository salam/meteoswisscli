package api

type RadarType string

const (
	RadarRain      RadarType = "rain"
	RadarCloud     RadarType = "cloud"
	RadarSatellite RadarType = "satellite"
)

var radarBrowserURLs = map[RadarType]string{
	RadarRain:      "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/niederschlag.html",
	RadarCloud:     "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/satellitenbilder.html#tab=satellite-animation-hrv",
	RadarSatellite: "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/satellitenbilder.html#tab=satellite-animation-hrv",
}

func GetRadarBrowserURL(rt RadarType) string { return radarBrowserURLs[rt] }
