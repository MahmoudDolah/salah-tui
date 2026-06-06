package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MahmoudDolah/salah-tui/internal/api"
	"github.com/MahmoudDolah/salah-tui/internal/cache"
	"github.com/MahmoudDolah/salah-tui/internal/config"
	"github.com/MahmoudDolah/salah-tui/internal/model"
	"github.com/MahmoudDolah/salah-tui/internal/prayer"
)

func main() {
	minimal := flag.Bool("minimal", false, "Display next prayer and countdown, then exit")
	watch := flag.Bool("watch", false, "In minimal mode, refresh every second instead of exiting")
	flag.Parse()

	// 1. Load config (runs first-run setup if missing)
	cfg, err := config.Load()
	if err != nil {
		// Config is malformed — re-run setup by deleting the file and retrying
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please fix or delete %s and re-run.\n", config.DefaultConfigPath())
		os.Exit(1)
	}

	// 2. Cleanup stale cache files
	today := time.Now()
	if err := cache.Cleanup(today); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cache cleanup failed: %v\n", err)
	}

	// 3. Load prayer times (cache → API → write cache)
	pt, err := loadPrayerTimes(cfg, today)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading prayer times: %v\n", err)
		os.Exit(1)
	}

	// 4. Load ayah (cache → API → write cache)
	ayah, err := loadAyah(today)
	if err != nil {
		// Non-fatal: ayah failure shouldn't block prayer times display
		fmt.Fprintf(os.Stderr, "Warning: could not load ayah: %v\n", err)
	}

	// 5. Resolve timezone location
	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: unknown timezone %q, defaulting to UTC: %v\n", cfg.Location.Timezone, err)
		loc = time.UTC
	}

	// 6. Parse prayer times into structured types
	prayers, err := prayer.ParseTimes(pt, today, loc, cfg.Display.ShowSunrise)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing prayer times: %v\n", err)
		os.Exit(1)
	}

	use12h := cfg.Display.TimeFormat == 12

	if *minimal {
		runMinimal(cfg, prayers, ayah, loc, use12h, *watch)
		return
	}

	// Dashboard mode
	offline := false
	m := model.New(cfg, loc, prayers, pt.HijriDate, pt.GregorianDate, ayah, offline)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running dashboard: %v\n", err)
		os.Exit(1)
	}
}

// runMinimal prints minimal output (next prayer + countdown) and either exits or loops.
// In watch mode, prayer times are reloaded automatically when the date rolls over midnight.
func runMinimal(cfg *config.Config, prayers []prayer.Prayer, ayah *api.Ayah, loc *time.Location, use12h, watch bool) {
	loadedDate := time.Now().In(loc)

	printMinimal := func() {
		now := time.Now().In(loc)

		if watch && (now.Year() != loadedDate.Year() || now.YearDay() != loadedDate.YearDay()) {
			loadedDate = now
			if newPrayers, err := loadAndParsePrayers(cfg, now, loc); err == nil {
				prayers = newPrayers
			} else {
				fmt.Fprintf(os.Stderr, "Warning: could not reload prayer times: %v\n", err)
			}
		}

		next := prayer.Next(prayers, now)
		if next == nil {
			fmt.Println("All prayers completed for today")
			return
		}
		d := prayer.Countdown(*next, now)
		fmt.Printf("%s (%s) in %s\n", next.Name, prayer.FormatTime(next.Time, use12h), prayer.FormatCountdown(d))
	}

	if !watch {
		printMinimal()
		return
	}

	// Watch mode: refresh every second
	for {
		printMinimal()
		time.Sleep(1 * time.Second)
	}
}

// loadAndParsePrayers fetches prayer times for the given date and parses them into Prayer structs.
func loadAndParsePrayers(cfg *config.Config, date time.Time, loc *time.Location) ([]prayer.Prayer, error) {
	pt, err := loadPrayerTimes(cfg, date)
	if err != nil {
		return nil, err
	}
	return prayer.ParseTimes(pt, date, loc, cfg.Display.ShowSunrise)
}

// loadPrayerTimes loads prayer times from cache if available, otherwise fetches from API.
func loadPrayerTimes(cfg *config.Config, today time.Time) (*api.PrayerTimes, error) {
	// Try cache first
	cached, err := cache.ReadPrayerTimes(today)
	if err != nil {
		return nil, fmt.Errorf("read prayer times cache: %w", err)
	}

	if cached != nil {
		var pt api.PrayerTimes
		if err := json.Unmarshal(cached, &pt); err == nil {
			return &pt, nil
		}
		// Malformed cache — fall through to fetch
	}

	// Fetch from API
	methodCode, ok := config.MethodCodes[cfg.Calculation.Method]
	if !ok {
		methodCode = config.MethodCodes["ISNA"]
	}

	pt, err := api.Fetch(cfg.Location.Latitude, cfg.Location.Longitude, methodCode)
	if err != nil {
		return nil, fmt.Errorf("fetch prayer times: %w", err)
	}

	// Write to cache
	data, err := json.Marshal(pt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not marshal prayer times for cache: %v\n", err)
	} else {
		if err := cache.WritePrayerTimes(today, data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write prayer times cache: %v\n", err)
		}
	}

	return pt, nil
}

// loadAyah loads the ayah of the day from cache if available, otherwise fetches from API.
func loadAyah(today time.Time) (*api.Ayah, error) {
	// Try cache first
	cached, err := cache.ReadAyah(today)
	if err != nil {
		return nil, fmt.Errorf("read ayah cache: %w", err)
	}

	if cached != nil {
		var ayah api.Ayah
		if err := json.Unmarshal(cached, &ayah); err == nil {
			return &ayah, nil
		}
		// Malformed cache — fall through to fetch
	}

	// Fetch from API using daily index
	index := api.DailyIndex(today)
	ayah, err := api.FetchByIndex(index)
	if err != nil {
		return nil, fmt.Errorf("fetch ayah: %w", err)
	}

	// Write to cache
	data, err := json.Marshal(ayah)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not marshal ayah for cache: %v\n", err)
	} else {
		if err := cache.WriteAyah(today, data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write ayah cache: %v\n", err)
		}
	}

	return ayah, nil
}
