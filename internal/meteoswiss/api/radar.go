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

var radarImageURLs = map[RadarType]string{
	RadarRain:      "https://www.meteoschweiz.admin.ch/product/output/radar/precip/animation/radar_precip.png",
	RadarCloud:     "https://www.meteoschweiz.admin.ch/product/output/satellite/animation/satellite_hrv.png",
	RadarSatellite: "https://www.meteoschweiz.admin.ch/product/output/satellite/animation/satellite_hrv.png",
}

func GetRadarBrowserURL(rt RadarType) string { return radarBrowserURLs[rt] }
func GetRadarImageURL(rt RadarType) string   { return radarImageURLs[rt] }
