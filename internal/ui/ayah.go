package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/MahmoudDolah/salah-tui/internal/api"
)

// RenderAyah renders the right pane: Arabic text, translation, and reference.
// width controls text wrapping for the translation.
func RenderAyah(ayah *api.Ayah, width int) string {
	if ayah == nil {
		return StyleError.Render("Ayah unavailable")
	}

	var b strings.Builder

	b.WriteString(StyleHeader.Render("Ayah of the Day") + "\n\n")

	// Arabic text — right-align within the pane
	arabicStyle := StyleAyahArabic.Width(width).Align(lipgloss.Right)
	b.WriteString(arabicStyle.Render(ayah.Arabic) + "\n\n")

	// English translation — wrapped
	translation := wrapText(ayah.Translation, width)
	b.WriteString(StyleAyahTranslation.Render(translation) + "\n\n")

	b.WriteString(StyleAyahReference.Render("— " + ayah.Reference))

	return b.String()
}

// RenderAyahLookup renders the right pane in lookup mode.
// query is what the user has typed so far; result is non-nil when a lookup completed.
func RenderAyahLookup(query string, result *api.Ayah, errMsg string, width int) string {
	var b strings.Builder

	b.WriteString(StyleHeader.Render("Quran Lookup") + "\n\n")
	b.WriteString(StyleDimText.Render("Enter surah:ayah (e.g. 2:255)") + "\n")
	b.WriteString(StyleNextPrayerLabel.Render("> " + query + "█") + "\n\n")

	switch {
	case errMsg != "":
		b.WriteString(StyleError.Render(errMsg))
	case result != nil:
		arabicStyle := StyleAyahArabic.Width(width).Align(lipgloss.Right)
		b.WriteString(arabicStyle.Render(result.Arabic) + "\n\n")
		b.WriteString(StyleAyahTranslation.Render(wrapText(result.Translation, width)) + "\n\n")
		b.WriteString(StyleAyahReference.Render("— " + result.Reference))
	}

	return b.String()
}

// wrapText soft-wraps s to at most maxWidth runes per line.
func wrapText(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) <= maxWidth {
			line += " " + w
		} else {
			lines = append(lines, line)
			line = w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
