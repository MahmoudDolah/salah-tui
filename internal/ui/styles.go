package ui

import "github.com/charmbracelet/lipgloss"

const (
	colorGold    = "#C9A84C"
	colorGreen   = "#4CAF50"
	colorCyan    = "#00BCD4"
	colorWhite   = "#FFFFFF"
	colorAmber   = "#FFA726"
	colorRed     = "#EF5350"
	colorDimWhite = "#AAAAAA"
	colorBorder  = "#444444"
)

// Countdown urgency thresholds in minutes.
const (
	ThresholdAmber = 15
	ThresholdRed   = 5
)

var (
	StyleHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorGold)).
		Bold(true)

	StyleCurrentPrayer = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorGreen)).
		Bold(true)

	StyleNextPrayerLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorCyan))

	StyleCountdownNormal = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorWhite))

	StyleCountdownAmber = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorAmber))

	StyleCountdownRed = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorRed))

	StyleAyahArabic = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorWhite)).
		Bold(true)

	StyleAyahTranslation = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorDimWhite))

	StyleAyahReference = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorGold))

	StyleBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder))

	StyleDimText = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorDimWhite))

	StyleHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorDimWhite))

	StyleError = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorRed))

	StyleOffline = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorAmber))
)

// CountdownStyle returns the appropriate countdown style based on duration in minutes.
func CountdownStyle(minutes float64) lipgloss.Style {
	switch {
	case minutes <= ThresholdRed:
		return StyleCountdownRed
	case minutes <= ThresholdAmber:
		return StyleCountdownAmber
	default:
		return StyleCountdownNormal
	}
}
