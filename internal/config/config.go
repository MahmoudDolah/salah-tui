package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// MethodCodes maps calculation method names to Aladhan API integer codes.
var MethodCodes = map[string]int{
	"ISNA":      2,
	"MWL":       3,
	"Egyptian":  5,
	"UmmAlQura": 4,
	"Karachi":   1,
	"Tehran":    7,
	"Shia":      0,
}

// Config is the top-level configuration structure.
type Config struct {
	Location    LocationConfig    `toml:"location"`
	Calculation CalculationConfig `toml:"calculation"`
	Display     DisplayConfig     `toml:"display"`
}

// LocationConfig holds geographic coordinates and timezone.
type LocationConfig struct {
	Latitude  float64 `toml:"latitude"`
	Longitude float64 `toml:"longitude"`
	City      string  `toml:"city"`
	Timezone  string  `toml:"timezone"`
}

// CalculationConfig holds the prayer calculation method.
type CalculationConfig struct {
	Method string `toml:"method"` // "ISNA", "MWL", "Egyptian", "UmmAlQura"
}

// DisplayConfig holds display preferences.
type DisplayConfig struct {
	TimeFormat  int  `toml:"time_format"`  // 12 or 24
	ShowSunrise bool `toml:"show_sunrise"`
}

// DefaultConfigPath returns the default path to the config file.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "salah", "config.toml")
	}
	return filepath.Join(home, ".config", "salah", "config.toml")
}

// Load reads the config file, running first-run setup if it doesn't exist.
// If the config is malformed, an error is returned.
func Load() (*Config, error) {
	path := DefaultConfigPath()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return runFirstRunSetup(path)
	}
	if err != nil {
		return nil, fmt.Errorf("stat config file: %w", err)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("malformed config file at %s: %w", path, err)
	}

	// Validate required fields
	if cfg.Location.Timezone == "" {
		cfg.Location.Timezone = "UTC"
	}
	if cfg.Calculation.Method == "" {
		cfg.Calculation.Method = "ISNA"
	}
	if cfg.Display.TimeFormat == 0 {
		cfg.Display.TimeFormat = 12
	}

	return &cfg, nil
}

// runFirstRunSetup prompts the user for config values and writes the config file.
func runFirstRunSetup(path string) (*Config, error) {
	fmt.Println("Welcome to salah-tui! Let's set up your configuration.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Prompt for city
	fmt.Print("Enter your city name: ")
	cityInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read city: %w", err)
	}
	city := strings.TrimSpace(cityInput)

	// Geocode the city
	lat, lng, timezone, err := geocodeCity(city)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not geocode city %q: %v\n", city, err)
		fmt.Println("Please enter coordinates manually.")

		fmt.Print("Enter latitude: ")
		latStr, _ := reader.ReadString('\n')
		latStr = strings.TrimSpace(latStr)
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude: %w", err)
		}

		fmt.Print("Enter longitude: ")
		lngStr, _ := reader.ReadString('\n')
		lngStr = strings.TrimSpace(lngStr)
		lng, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude: %w", err)
		}

		timezone = "UTC"
	}

	fmt.Printf("Located %s at (%.4f, %.4f), timezone: %s\n", city, lat, lng, timezone)
	fmt.Println()

	// Prompt for calculation method
	fmt.Println("Select calculation method:")
	fmt.Println("  1. ISNA (Islamic Society of North America) [default]")
	fmt.Println("  2. MWL (Muslim World League)")
	fmt.Println("  3. Egyptian (Egyptian General Authority of Survey)")
	fmt.Println("  4. UmmAlQura (Umm Al-Qura University, Makkah)")
	fmt.Print("Enter choice [1-4, default 1]: ")

	methodInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read method choice: %w", err)
	}
	methodInput = strings.TrimSpace(methodInput)

	method := "ISNA"
	switch methodInput {
	case "", "1":
		method = "ISNA"
	case "2":
		method = "MWL"
	case "3":
		method = "Egyptian"
	case "4":
		method = "UmmAlQura"
	default:
		fmt.Fprintf(os.Stderr, "Unknown choice %q, defaulting to ISNA\n", methodInput)
	}

	cfg := &Config{
		Location: LocationConfig{
			Latitude:  lat,
			Longitude: lng,
			City:      city,
			Timezone:  timezone,
		},
		Calculation: CalculationConfig{
			Method: method,
		},
		Display: DisplayConfig{
			TimeFormat:  12,
			ShowSunrise: true,
		},
	}

	if err := writeConfig(cfg, path); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("Config saved to %s\n\n", path)
	return cfg, nil
}

// writeConfig serializes the config struct to TOML and writes it to path.
func writeConfig(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	content := fmt.Sprintf(`[location]
latitude = %f
longitude = %f
city = %q
timezone = %q

[calculation]
# Options: ISNA, MWL, Egyptian, Karachi, UmmAlQura, Tehran, Shia
method = %q

[display]
# 12 or 24
time_format = %d
# Show Sunrise in the schedule (not a prayer, but useful)
show_sunrise = %v
`,
		cfg.Location.Latitude,
		cfg.Location.Longitude,
		cfg.Location.City,
		cfg.Location.Timezone,
		cfg.Calculation.Method,
		cfg.Display.TimeFormat,
		cfg.Display.ShowSunrise,
	)

	_, err = f.WriteString(content)
	return err
}

// nominatimResult is the JSON structure returned by Nominatim geocoding.
type nominatimResult struct {
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Display string `json:"display_name"`
}

// geocodeCity looks up lat/lng for a city name using Nominatim.
// Returns lat, lng, timezone (always "UTC" since Nominatim doesn't provide timezone).
func geocodeCity(city string) (lat, lng float64, timezone string, err error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := "https://nominatim.openstreetmap.org/search?q=" +
		strings.ReplaceAll(city, " ", "+") +
		"&format=json&limit=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "salah-tui/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, "", fmt.Errorf("geocode request: %w", err)
	}
	defer resp.Body.Close()

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, "", fmt.Errorf("decode geocode response: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, "", fmt.Errorf("no results found for city %q", city)
	}

	lat, err = strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("parse latitude %q: %w", results[0].Lat, err)
	}

	lng, err = strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("parse longitude %q: %w", results[0].Lon, err)
	}

	// Nominatim does not return timezone; default to UTC
	return lat, lng, "UTC", nil
}
