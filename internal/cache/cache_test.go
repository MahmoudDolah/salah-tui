package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var today = time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC)
var yesterday = today.AddDate(0, 0, -1)

func TestWriteReadPrayerTimes(t *testing.T) {
	data := []byte(`{"Fajr":"05:30"}`)
	if err := WritePrayerTimes(today, data); err != nil {
		t.Fatalf("WritePrayerTimes: %v", err)
	}
	t.Cleanup(func() {
		dir, _ := cacheDir()
		os.Remove(filepath.Join(dir, prayerTimesFilename(today)))
	})

	got, err := ReadPrayerTimes(today)
	if err != nil {
		t.Fatalf("ReadPrayerTimes: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestWriteReadAyah(t *testing.T) {
	data := []byte(`{"Arabic":"بسم الله","Translation":"In the name of God"}`)
	if err := WriteAyah(today, data); err != nil {
		t.Fatalf("WriteAyah: %v", err)
	}
	t.Cleanup(func() {
		dir, _ := cacheDir()
		os.Remove(filepath.Join(dir, ayahFilename(today)))
	})

	got, err := ReadAyah(today)
	if err != nil {
		t.Fatalf("ReadAyah: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestReadPrayerTimes_NotFound(t *testing.T) {
	// Use a date far in the future that will never have a cache file.
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := ReadPrayerTimes(future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing cache, got %q", got)
	}
}

func TestReadAyah_NotFound(t *testing.T) {
	future := time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC)
	got, err := ReadAyah(future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing cache, got %q", got)
	}
}

func TestCleanup(t *testing.T) {
	// Write today + yesterday files.
	todayPT := []byte(`{"today":"prayer"}`)
	todayAyah := []byte(`{"today":"ayah"}`)
	yestPT := []byte(`{"yesterday":"prayer"}`)
	yestAyah := []byte(`{"yesterday":"ayah"}`)

	for _, write := range []func() error{
		func() error { return WritePrayerTimes(today, todayPT) },
		func() error { return WriteAyah(today, todayAyah) },
		func() error { return WritePrayerTimes(yesterday, yestPT) },
		func() error { return WriteAyah(yesterday, yestAyah) },
	} {
		if err := write(); err != nil {
			t.Fatalf("setup write: %v", err)
		}
	}

	dir, err := cacheDir()
	if err != nil {
		t.Fatalf("cacheDir: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(filepath.Join(dir, prayerTimesFilename(today)))
		os.Remove(filepath.Join(dir, ayahFilename(today)))
		// yesterday files should be gone; remove just in case
		os.Remove(filepath.Join(dir, prayerTimesFilename(yesterday)))
		os.Remove(filepath.Join(dir, ayahFilename(yesterday)))
	})

	if err := Cleanup(today); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	// Yesterday's files must be gone.
	for _, name := range []string{prayerTimesFilename(yesterday), ayahFilename(yesterday)} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed by Cleanup", name)
		}
	}

	// Today's files must still exist.
	for _, name := range []string{prayerTimesFilename(today), ayahFilename(today)} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to still exist after Cleanup: %v", name, err)
		}
	}
}
