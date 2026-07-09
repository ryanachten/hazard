package tui

import (
	c "hazard/internal/common"

	"charm.land/lipgloss/v2"
)

var citizenCharacter = "@"
var citizenStyle = lipgloss.NewStyle().
	SetString(citizenCharacter).
	Foreground(lipgloss.Color("#eab308")).
	Background(lipgloss.Color("#2a2000")).Render()

var citizenDeadCharacter = "†"
var citizenDeadStyle = lipgloss.NewStyle().
	SetString(citizenDeadCharacter).
	Foreground(lipgloss.Color("#991b1b")).
	Background(lipgloss.Color("#2a0505")).
	Render()

var citizenEscapedCharacter = "♦"
var citizenEscapedStyle = lipgloss.NewStyle().
	SetString(citizenEscapedCharacter).
	Foreground(lipgloss.Color("#4ade80")).
	Background(lipgloss.Color("#0d2a0d")).
	Render()

var openCharacters = []string{"·", ".", ",", " "}
var openStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#586069")).
	Background(lipgloss.Color("#0d1117"))

var obstacleCharacters = []string{"#", "▓", "▒", "≡"}
var obstacleStyle = lipgloss.NewStyle().Inherit(openStyle)

var fireCharacters = []string{"░", "\"", "^"}
var fireStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#ef4444")).
	Background(lipgloss.Color("#2a0a0a"))

var floodCharacters = []string{"~", "≈", "░"}
var floodStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#3b82f6")).
	Background(lipgloss.Color("#0a1628"))

var lavaCharacters = []string{"≈", "▒", "°", "o"}
var lavaStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#f97316")).
	Background(lipgloss.Color("#2a1400"))

var safeZoneCharacters = []string{"◎", "◉", "◌"}
var safeZoneStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#22c55e")).
	Background(lipgloss.Color("#0d2a0d"))

func getOpenCell() string {
	char := c.RandValInSlice(openCharacters)
	bgOffset := c.RandomFloat(0, 0.04)
	bgColour := lipgloss.Color("#0d1117")

	return openStyle.
		SetString(char).
		Background(lipgloss.Lighten(bgColour, bgOffset)).
		Render()
}

func getCitizenCell() string {
	return citizenStyle
}

func getDeadCitizenCell() string {
	return citizenDeadStyle
}

func getEscapedCitizenCell() string {
	return citizenEscapedStyle
}

func getFireCell() string {
	char := c.RandValInSlice(fireCharacters)
	return fireStyle.SetString(char).Render()
}

func getFloodCell() string {
	char := c.RandValInSlice(floodCharacters)
	return floodStyle.SetString(char).Render()
}

func getLavaCell() string {
	char := c.RandValInSlice(lavaCharacters)
	return lavaStyle.SetString(char).Render()
}

func getSafeZoneCell(char string) string {
	return safeZoneStyle.SetString(char).Render()
}

func getObstacleCell() string {
	char := c.RandValInSlice(obstacleCharacters)
	return obstacleStyle.SetString(char).Render()
}

var gridStyle = lipgloss.NewStyle().MarginRight(2)
