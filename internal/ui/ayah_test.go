package ui

import (
	"strings"
	"testing"

	"github.com/MahmoudDolah/salah-tui/internal/api"
)

var testAyah = &api.Ayah{
	Arabic:      "إِنَّا أَعْطَيْنَاكَ الْكَوْثَرَ",
	Translation: "Verily We have granted you the Abundance",
	Reference:   "Al-Kawthar 108:1",
	Surah:       108,
	AyahNum:     1,
}

func TestRenderAyah_NilReturnsUnavailable(t *testing.T) {
	out := stripANSI(RenderAyah(nil, "", 40))
	if !strings.Contains(out, "unavailable") {
		t.Errorf("expected 'unavailable' for nil ayah, got: %q", out)
	}
}

func TestRenderAyah_NilWithErrorShowsErrorAndRetry(t *testing.T) {
	out := stripANSI(RenderAyah(nil, "fetch ayah: timeout", 40))
	if !strings.Contains(out, "fetch ayah: timeout") {
		t.Errorf("expected error message in ayah pane, got: %q", out)
	}
	if !strings.Contains(out, "Press r to retry") {
		t.Errorf("expected retry hint in ayah pane, got: %q", out)
	}
}

func TestRenderAyah_NilWithErrorShowsHeader(t *testing.T) {
	out := stripANSI(RenderAyah(nil, "some error", 40))
	if !strings.Contains(out, "Ayah of the Day") {
		t.Errorf("expected header in ayah error pane, got: %q", out)
	}
}

func TestRenderAyah_ContainsArabic(t *testing.T) {
	out := stripANSI(RenderAyah(testAyah, "", 40))
	if !strings.Contains(out, testAyah.Arabic) {
		t.Error("expected Arabic text in output")
	}
}

func TestRenderAyah_ContainsTranslation(t *testing.T) {
	out := stripANSI(RenderAyah(testAyah, "", 40))
	// Translation may be wrapped; check for a substring
	if !strings.Contains(out, "Verily We have granted") {
		t.Errorf("expected translation text in output, got:\n%s", out)
	}
}

func TestRenderAyah_ContainsReference(t *testing.T) {
	out := stripANSI(RenderAyah(testAyah, "", 40))
	if !strings.Contains(out, "Al-Kawthar 108:1") {
		t.Error("expected reference in output")
	}
}

func TestRenderAyahLookup_ShowsQuery(t *testing.T) {
	out := stripANSI(RenderAyahLookup("2:25", nil, "", 40))
	if !strings.Contains(out, "2:25") {
		t.Errorf("expected query in lookup output, got: %q", out)
	}
}

func TestRenderAyahLookup_ShowsError(t *testing.T) {
	out := stripANSI(RenderAyahLookup("bad", nil, "invalid reference", 40))
	if !strings.Contains(out, "invalid reference") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}

func TestRenderAyahLookup_ShowsResult(t *testing.T) {
	out := stripANSI(RenderAyahLookup("108:1", testAyah, "", 40))
	if !strings.Contains(out, testAyah.Arabic) {
		t.Error("expected Arabic text in lookup result")
	}
	if !strings.Contains(out, "Al-Kawthar 108:1") {
		t.Error("expected reference in lookup result")
	}
}

func TestRenderAyahLookup_ErrorTakesPrecedenceOverResult(t *testing.T) {
	// When both err and result are set, errMsg is shown (switch case order)
	out := stripANSI(RenderAyahLookup("108:1", testAyah, "some error", 40))
	if !strings.Contains(out, "some error") {
		t.Error("expected error message to be shown")
	}
}

func TestWrapText_ShortTextUnchanged(t *testing.T) {
	got := wrapText("Hello world", 40)
	if got != "Hello world" {
		t.Errorf("got %q, want %q", got, "Hello world")
	}
}

func TestWrapText_WrapsAtWidth(t *testing.T) {
	got := wrapText("one two three four five", 10)
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if len(line) > 10 {
			t.Errorf("line %q exceeds width 10", line)
		}
	}
}

func TestWrapText_EmptyString(t *testing.T) {
	got := wrapText("", 40)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestWrapText_ZeroWidth(t *testing.T) {
	s := "Hello world"
	got := wrapText(s, 0)
	if got != s {
		t.Errorf("zero width should return original string, got %q", got)
	}
}

func TestWrapText_SingleLongWord(t *testing.T) {
	// A single word longer than width should not be split (no hyphenation)
	got := wrapText("superlongword", 5)
	if got != "superlongword" {
		t.Errorf("single word should not be broken, got %q", got)
	}
}
