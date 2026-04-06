package api

import "fmt"

type BulletinResponse struct {
	Bulletins  []Bulletin `json:"bulletins"`
	CustomData any        `json:"customData,omitempty"`
}

type Bulletin struct {
	BulletinID        string             `json:"bulletinID"`
	Lang              string             `json:"lang"`
	PublicationTime   string             `json:"publicationTime"`
	ValidTime         ValidTime          `json:"validTime"`
	NextUpdate        string             `json:"nextUpdate,omitempty"`
	Unscheduled       bool               `json:"unscheduled,omitempty"`
	Regions           []Region           `json:"regions"`
	DangerRatings     []DangerRating     `json:"dangerRatings,omitempty"`
	AvalancheProblems []AvalancheProblem `json:"avalancheProblems,omitempty"`
	WeatherForecast   *TextContent       `json:"weatherForecast,omitempty"`
	SnowpackStructure *TextContent       `json:"snowpackStructure,omitempty"`
	TravelAdvisory    *TextContent       `json:"travelAdvisory,omitempty"`
	Tendency          []Tendency         `json:"tendency,omitempty"`
}

type ValidTime struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type Region struct {
	RegionID string `json:"regionID"`
	Name     string `json:"name"`
}

type DangerRating struct {
	MainValue       string         `json:"mainValue"`
	Elevation       ElevationRange `json:"elevation,omitempty"`
	ValidTimePeriod string         `json:"validTimePeriod,omitempty"`
}

type ElevationRange struct {
	LowerBound string `json:"lowerBound,omitempty"`
	UpperBound string `json:"upperBound,omitempty"`
}

type AvalancheProblem struct {
	ProblemType     string         `json:"problemType"`
	Elevation       ElevationRange `json:"elevation,omitempty"`
	Aspects         []string       `json:"aspects,omitempty"`
	ValidTimePeriod string         `json:"validTimePeriod,omitempty"`
}

type TextContent struct {
	Comment string `json:"comment,omitempty"`
}

type Tendency struct {
	Comment      string     `json:"comment,omitempty"`
	TendencyType string     `json:"tendencyType,omitempty"`
	ValidTime    *ValidTime `json:"validTime,omitempty"`
}

func (c *Client) GetBulletin() (*BulletinResponse, error) {
	url := fmt.Sprintf("%s/api/bulletin/caaml/%s/json", c.bulletinBase, c.lang)
	var result BulletinResponse
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get bulletin: %w", err)
	}
	return &result, nil
}

func (c *Client) GetBulletinPDFURL() string {
	return fmt.Sprintf("%s/api/bulletin/document/full/%s", c.bulletinBase, c.lang)
}

var dangerLevelNames = map[string]string{
	"low":          "1 — Low",
	"moderate":     "2 — Moderate",
	"considerable": "3 — Considerable",
	"high":         "4 — High",
	"very_high":    "5 — Very High",
}

func DangerLevelDisplay(value string) string {
	if name, ok := dangerLevelNames[value]; ok {
		return name
	}
	return value
}
