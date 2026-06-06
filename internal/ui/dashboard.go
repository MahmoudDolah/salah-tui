package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/MahmoudDolah/salah-tui/internal/api"
	"github.com/MahmoudDolah/salah-tui/internal/prayer"
)

const minDashboardWidth = 60

// RenderDashboard composes the full two-pane dashboard view.
func RenderDashboard(
	prayers []prayer.Prayer,
	hijriDate, gregorianDate string,
	ayah *api.Ayah,
	now time.Time,
	use12h bool,
	offline bool,
	lookupMode bool,
	lookupQuery string,
	lookupResult *api.Ayah,
	lookupErr string,
	showHelp bool,
	termWidth, termHeight int,
) string {
	if termWidth < minDashboardWidth {
		return renderNarrow(prayers, now, use12h, termWidth)
	}

	borderW := 2 // one rounded border adds 2 cols (left+right)
	gap := 2
	leftWidth := 36
	rightWidth := termWidth - leftWidth - gap - borderW*2
	if rightWidth < 20 {
		rightWidth = 20
	}

	leftContent := RenderSchedule(prayers, hijriDate, gregorianDate, now, use12h, offline, leftWidth)
	var rightContent string
	if lookupMode {
		rightContent = RenderAyahLookup(lookupQuery, lookupResult, lookupErr, rightWidth)
	} else {
		rightContent = RenderAyah(ayah, rightWidth)
	}

	leftPane := StyleBorder.Width(leftWidth).Height(termHeight - 4).Render(leftContent)
	rightPane := StyleBorder.Width(rightWidth).Height(termHeight - 4).Render(rightContent)

	dashboard := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, strings.Repeat(" ", gap), rightPane)

	if showHelp {
		return lipgloss.JoinVertical(lipgloss.Left, dashboard, renderHelp())
	}
	return dashboard
}

// renderNarrow is a minimal fallback for terminals that are too narrow for the split layout.
func renderNarrow(prayers []prayer.Prayer, now time.Time, use12h bool, width int) string {
	next := prayer.Next(prayers, now)
	notice := StyleOffline.Render("(terminal too narrow for dashboard)") + "\n\n"
	if next == nil {
		return notice + "All prayers completed for today."
	}
	d := prayer.Countdown(*next, now)
	mins := d.Minutes()
	timeStr := prayer.FormatTime(next.Time, use12h)
	_ = width
	return notice +
		StyleNextPrayerLabel.Render("Next: "+next.Name) + "  " +
		StyleDimText.Render(timeStr) + "\n" +
		CountdownStyle(mins).Render(prayer.FormatCountdown(d))
}

// renderHelp returns the help overlay string.
func renderHelp() string {
	rows := []struct{ key, desc string }{
		{"q / ctrl+c", "Quit"},
		{"/", "Open Quran lookup"},
		{"esc", "Close lookup"},
		{"r", "Force refresh"},
		{"?", "Toggle this help"},
	}
	var b strings.Builder
	b.WriteString("\n" + StyleHeader.Render("Keybindings") + "\n")
	for _, r := range rows {
		b.WriteString(StyleNextPrayerLabel.Render(r.key) + "\t" + StyleHelp.Render(r.desc) + "\n")
	}
	return b.String()
}
