package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// cacheDir returns the salah cache directory path.
func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".cache", "salah"), nil
}

// ensureCacheDir creates the cache directory if it does not exist.
func ensureCacheDir() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}
	return dir, nil
}

// prayerTimesFilename returns the cache filename for prayer times on the given date.
func prayerTimesFilename(date time.Time) string {
	return "prayer-times-" + date.Format("2006-01-02") + ".json"
}

// ayahFilename returns the cache filename for the ayah on the given date.
func ayahFilename(date time.Time) string {
	return "ayah-" + date.Format("2006-01-02") + ".json"
}

// ReadPrayerTimes reads today's prayer times cache.
// Returns nil, nil if not found.
func ReadPrayerTimes(date time.Time) ([]byte, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, prayerTimesFilename(date))
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read prayer times cache: %w", err)
	}
	return data, nil
}

// WritePrayerTimes writes today's prayer times cache.
func WritePrayerTimes(date time.Time, data []byte) error {
	dir, err := ensureCacheDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, prayerTimesFilename(date))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write prayer times cache: %w", err)
	}
	return nil
}

// ReadAyah reads today's ayah cache.
// Returns nil, nil if not found.
func ReadAyah(date time.Time) ([]byte, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, ayahFilename(date))
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read ayah cache: %w", err)
	}
	return data, nil
}

// WriteAyah writes today's ayah cache.
func WriteAyah(date time.Time, data []byte) error {
	dir, err := ensureCacheDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, ayahFilename(date))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write ayah cache: %w", err)
	}
	return nil
}

// Cleanup removes cache files that are not from today.
func Cleanup(today time.Time) error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}

	// If the cache dir doesn't exist yet, nothing to clean up.
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read cache dir: %w", err)
	}

	todayStr := today.Format("2006-01-02")
	for _, entry := range entries {
		name := entry.Name()
		// Only touch files that match our naming patterns
		if !strings.HasPrefix(name, "prayer-times-") && !strings.HasPrefix(name, "ayah-") {
			continue
		}
		// Keep today's files
		if strings.Contains(name, todayStr) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale cache file %s: %w", name, err)
		}
	}
	return nil
}
