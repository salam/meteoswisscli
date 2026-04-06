package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type BulletinType string

const (
	BulletinReport  BulletinType = "weather-report"
	BulletinOutlook BulletinType = "weather-outlook"
)

type BulletinRegion string

const (
	RegionNorth BulletinRegion = "north"
	RegionSouth BulletinRegion = "south"
	RegionWest  BulletinRegion = "west"
)

type BulletinText struct {
	Type    BulletinType   `json:"type"`
	Region  BulletinRegion `json:"region"`
	Lang    string         `json:"lang"`
	Version string         `json:"version"`
	HTML    string         `json:"html,omitempty"`
	Text    string         `json:"text"`
}

var tagRe = regexp.MustCompile(`<[^>]+>`)
var spaceRe = regexp.MustCompile(`\s+`)

func htmlToText(html string) string {
	// Replace block elements with newlines
	for _, tag := range []string{"</h3>", "</h4>", "</p>", "<br>", "<br/>", "<br />"} {
		html = strings.ReplaceAll(html, tag, "\n")
	}
	// Add emphasis markers for headings
	html = strings.ReplaceAll(html, "<h3>", "\n## ")
	html = strings.ReplaceAll(html, "<h4>", "\n### ")
	// Strip remaining tags
	text := tagRe.ReplaceAllString(html, "")
	// Decode HTML entities
	text = strings.ReplaceAll(text, "&uuml;", "ü")
	text = strings.ReplaceAll(text, "&ouml;", "ö")
	text = strings.ReplaceAll(text, "&auml;", "ä")
	text = strings.ReplaceAll(text, "&Uuml;", "Ü")
	text = strings.ReplaceAll(text, "&Ouml;", "Ö")
	text = strings.ReplaceAll(text, "&Auml;", "Ä")
	text = strings.ReplaceAll(text, "&eacute;", "é")
	text = strings.ReplaceAll(text, "&egrave;", "è")
	text = strings.ReplaceAll(text, "&agrave;", "à")
	text = strings.ReplaceAll(text, "&ocirc;", "ô")
	text = strings.ReplaceAll(text, "&icirc;", "î")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&deg;", "°")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&laquo;", "«")
	text = strings.ReplaceAll(text, "&raquo;", "»")
	// Clean up whitespace
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

func (c *Client) GetBulletinText(bulletinType BulletinType, region BulletinRegion) (*BulletinText, error) {
	lang := c.lang
	if lang == "en" {
		lang = "de" // No English bulletins, fall back to German
	}

	// Get current version
	versionsURL := fmt.Sprintf("%s/product/output/%s/%s/%s/versions.json",
		"https://www.meteoschweiz.admin.ch", bulletinType, lang, region)

	var versions struct {
		CurrentVersionDirectory string `json:"currentVersionDirectory"`
	}
	data, err := c.DoRaw("GET", versionsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch bulletin version: %w", err)
	}
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("parse bulletin version: %w", err)
	}

	// Fetch text content
	textURL := fmt.Sprintf("%s/product/output/%s/%s/%s/%s/textproduct_%s.xhtml",
		"https://www.meteoschweiz.admin.ch", bulletinType, lang, region,
		versions.CurrentVersionDirectory, lang)

	htmlData, err := c.DoRaw("GET", textURL)
	if err != nil {
		return nil, fmt.Errorf("fetch bulletin text: %w", err)
	}

	html := string(htmlData)
	return &BulletinText{
		Type:    bulletinType,
		Region:  region,
		Lang:    lang,
		Version: versions.CurrentVersionDirectory,
		HTML:    html,
		Text:    htmlToText(html),
	}, nil
}
