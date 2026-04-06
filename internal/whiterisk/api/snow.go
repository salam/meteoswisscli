package api

type SnowMapType string

const (
	SnowMapNew     SnowMapType = "new"
	SnowMapDepth   SnowMapType = "depth"
	SnowMapCompare SnowMapType = "compare"
)

var snowBrowserURLs = map[SnowMapType]string{
	SnowMapNew:     "https://whiterisk.ch/de/conditions/snow-maps/new_snow",
	SnowMapDepth:   "https://whiterisk.ch/de/conditions/snow-maps/snow_depth",
	SnowMapCompare: "https://whiterisk.ch/de/conditions/snow-maps/comparative_snow_depth",
}

var snowTeaserURLs = map[SnowMapType]string{
	SnowMapNew:     "https://whiterisk.ch/snowmap-teaser/new-snow.png",
	SnowMapDepth:   "https://whiterisk.ch/snowmap-teaser/snow-depth.png",
	SnowMapCompare: "https://whiterisk.ch/snowmap-teaser/comparative-snow-depth.png",
}

func GetSnowBrowserURL(t SnowMapType) string { return snowBrowserURLs[t] }
func GetSnowTeaserURL(t SnowMapType) string  { return snowTeaserURLs[t] }
