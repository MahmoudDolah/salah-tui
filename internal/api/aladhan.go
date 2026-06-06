package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PrayerTimes holds the parsed prayer times for a single day.
type PrayerTimes struct {
	Fajr    string
	Sunrise string
	Dhuhr   string
	Asr     string
	Maghrib string
	Isha    string

	// Date information
	HijriDate     string // e.g. "09 Dhū al-Ḥijjah 1446"
	GregorianDate string // e.g. "06 Jun 2025"
}

// aladhanResponse is the top-level JSON envelope from the Aladhan API.
type aladhanResponse struct {
	Code   int            `json:"code"`
	Status string         `json:"status"`
	Data   aladhanData    `json:"data"`
}

type aladhanData struct {
	Timings aladhanTimings `json:"timings"`
	Date    aladhanDate    `json:"date"`
}

type aladhanTimings struct {
	Fajr    string `json:"Fajr"`
	Sunrise string `json:"Sunrise"`
	Dhuhr   string `json:"Dhuhr"`
	Asr     string `json:"Asr"`
	Maghrib string `json:"Maghrib"`
	Isha    string `json:"Isha"`
}

type aladhanDate struct {
	Readable string      `json:"readable"`
	Hijri    hijriDate   `json:"hijri"`
}

type hijriDate struct {
	Date  string     `json:"date"`
	Month hijriMonth `json:"month"`
	Year  string     `json:"year"`
}

type hijriMonth struct {
	En string `json:"en"`
}

var aladhanClient = &http.Client{Timeout: 10 * time.Second}

// Fetch retrieves prayer times from the Aladhan API for today's date.
func Fetch(lat, lng float64, methodCode int) (*PrayerTimes, error) {
	now := time.Now()
	ts := now.Unix()

	url := fmt.Sprintf(
		"https://api.aladhan.com/v1/timings/%d?latitude=%f&longitude=%f&method=%d",
		ts, lat, lng, methodCode,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build aladhan request: %w", err)
	}
	req.Header.Set("User-Agent", "salah-tui/0.1")

	resp, err := aladhanClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aladhan request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aladhan API returned status %d", resp.StatusCode)
	}

	var result aladhanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode aladhan response: %w", err)
	}

	t := result.Data.Timings
	d := result.Data.Date

	// Build the Hijri date string: "09 Dhū al-Ḥijjah 1446"
	// d.Hijri.Date is "DD-MM-YYYY" from the API; take only the day part.
	hijriStr := d.Hijri.Date[:2] + " " + d.Hijri.Month.En + " " + d.Hijri.Year
	// The date field from API is "DD-MM-YYYY", reformat for display
	// We prefer d.Readable for the Gregorian date as it's already human-friendly
	pt := &PrayerTimes{
		Fajr:          stripTimezone(t.Fajr),
		Sunrise:       stripTimezone(t.Sunrise),
		Dhuhr:         stripTimezone(t.Dhuhr),
		Asr:           stripTimezone(t.Asr),
		Maghrib:       stripTimezone(t.Maghrib),
		Isha:          stripTimezone(t.Isha),
		HijriDate:     hijriStr,
		GregorianDate: d.Readable,
	}

	return pt, nil
}

// stripTimezone removes any trailing timezone suffix (e.g. " (EDT)") from time strings
// returned by the Aladhan API, leaving only "HH:MM".
func stripTimezone(s string) string {
	for i, ch := range s {
		if ch == ' ' {
			return s[:i]
		}
	}
	return s
}
