package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Ayah holds a single Quranic verse with its Arabic text and English translation.
type Ayah struct {
	Arabic      string
	Translation string
	Reference   string // e.g. "Al-Baqarah 2:255"
	Surah       int
	AyahNum     int
}

// quranResponse is the JSON envelope returned by alquran.cloud for multi-edition requests.
type quranResponse struct {
	Code   int          `json:"code"`
	Status string       `json:"status"`
	Data   []quranAyah  `json:"data"`
}

type quranAyah struct {
	Number       int        `json:"number"`
	Text         string     `json:"text"`
	Surah        quranSurah `json:"surah"`
	NumberInSurah int       `json:"numberInSurah"`
}

type quranSurah struct {
	Number       int    `json:"number"`
	Name         string `json:"name"`         // Arabic name
	EnglishName  string `json:"englishName"`  // English name
}

var quranClient = &http.Client{Timeout: 10 * time.Second}

// FetchByIndex fetches an ayah by its global Quran index (1–6236).
func FetchByIndex(index int) (*Ayah, error) {
	url := fmt.Sprintf(
		"https://api.alquran.cloud/v1/ayah/%d/editions/quran-uthmani,en.asad",
		index,
	)
	return fetchAyah(url)
}

// FetchByRef fetches an ayah by surah and ayah number.
func FetchByRef(surah, ayah int) (*Ayah, error) {
	url := fmt.Sprintf(
		"https://api.alquran.cloud/v1/ayah/%d:%d/editions/quran-uthmani,en.asad",
		surah, ayah,
	)
	return fetchAyah(url)
}

// DailyIndex returns the global ayah index for the given date.
// Uses ((dayOfYear - 1) % 6236) + 1 to stay in 1–6236 range.
func DailyIndex(t time.Time) int {
	dayOfYear := t.YearDay()
	return ((dayOfYear - 1) % 6236) + 1
}

// fetchAyah performs the HTTP request and parses the response for the given URL.
func fetchAyah(url string) (*Ayah, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build quran request: %w", err)
	}
	req.Header.Set("User-Agent", "salah-tui/0.1")

	resp, err := quranClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quran request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("quran API returned status %d", resp.StatusCode)
	}

	var result quranResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode quran response: %w", err)
	}

	// We expect two editions: index 0 is Arabic (quran-uthmani), index 1 is English (en.asad)
	if len(result.Data) < 2 {
		return nil, fmt.Errorf("unexpected quran response: expected 2 editions, got %d", len(result.Data))
	}

	arabic := result.Data[0]
	english := result.Data[1]

	ref := fmt.Sprintf("%s %d:%d", arabic.Surah.EnglishName, arabic.Surah.Number, arabic.NumberInSurah)

	return &Ayah{
		Arabic:      arabic.Text,
		Translation: english.Text,
		Reference:   ref,
		Surah:       arabic.Surah.Number,
		AyahNum:     arabic.NumberInSurah,
	}, nil
}
