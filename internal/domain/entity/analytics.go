package entity

import "time"

// DailyAnalytics represents aggregated link statistics for a specific day.
type DailyAnalytics struct {
	LinkHash   string
	Date       time.Time
	Country    string
	Device     string
	Browser    string
	ClickCount int
	UniqueIPs  int
}

// AnalyticsReport is a structure for responding with aggregated stats
type AnalyticsReport struct {
	Hash           string             `json:"hash"`
	TotalClicks    int                `json:"total_clicks"`
	UniqueVisitors int                `json:"unique_visitors"`
	Period         Period             `json:"period"`
	ByCountry      []DimensionStat    `json:"by_country"`
	ByDevice       []DimensionStat    `json:"by_device"`
	ByBrowser      []DimensionStat    `json:"by_browser"`
	ByDay          []DailyStat        `json:"by_day"`
}

type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type DimensionStat struct {
	Name       string  `json:"name"` // Used dynamically for country/device/browser etc.
	Clicks     int     `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

type DailyStat struct {
	Date   string `json:"date"`
	Clicks int    `json:"clicks"`
}
