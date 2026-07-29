package tui

import (
	"charm.land/lipgloss/v2"

	"hazard/internal/random"
)

var citizenStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#eab308")).
	Background(lipgloss.Color("#2a2000"))
var citizenCharacter = lipgloss.NewStyle().
	Inherit(citizenStyle).
	SetString("@").
	Render()

var pathHorizontalCharacter = lipgloss.NewStyle().
	Inherit(citizenStyle).
	SetString("-").
	Render()
var pathVerticalCharacter = lipgloss.NewStyle().
	Inherit(citizenStyle).
	SetString("|").
	Render()

var citizenDeadStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#c41b1b")).
	Background(lipgloss.Color("#2a0505"))
var citizenDeadCharacter = lipgloss.NewStyle().
	Inherit(citizenDeadStyle).
	SetString("†").
	Render()

var citizenEscapedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#4ade80")).
	Background(lipgloss.Color("#0d2a0d"))
var citizenEscapedCharacter = lipgloss.NewStyle().
	Inherit(citizenEscapedStyle).
	SetString("♦").
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
	char := random.ValInSlice(openCharacters)
	bgOffset := random.Float(0, 0.04)
	bgColour := lipgloss.Color("#0d1117")

	return openStyle.
		SetString(char).
		Background(lipgloss.Lighten(bgColour, bgOffset)).
		Render()
}

func getFireCell() string {
	char := random.ValInSlice(fireCharacters)
	return fireStyle.SetString(char).Render()
}

func getFloodCell() string {
	char := random.ValInSlice(floodCharacters)
	return floodStyle.SetString(char).Render()
}

func getLavaCell() string {
	char := random.ValInSlice(lavaCharacters)
	return lavaStyle.SetString(char).Render()
}

func getSafeZoneCell(char string) string {
	return safeZoneStyle.SetString(char).Render()
}

func getObstacleCell() string {
	char := random.ValInSlice(obstacleCharacters)
	return obstacleStyle.SetString(char).Render()
}

var logoLines = []string{
	"░█░█░█▀█░▀▀█░█▀█░█▀▄░█▀▄",
	"░█▀█░█▀█░▄▀░░█▀█░█▀▄░█░█",
	"░▀░▀░▀░▀░▀▀▀░▀░▀░▀░▀░▀▀░",
}

var logoChars [][]string
var logoCharWidth, logoCharHeight int

func init() {
	logoChars = make([][]string, len(logoLines))
	for i, line := range logoLines {
		runes := []rune(line)
		chars := make([]string, len(runes))
		for j, r := range runes {
			chars[j] = string(r)
		}
		logoChars[i] = chars
	}
	logoCharHeight = len(logoChars)
	if logoCharHeight > 0 {
		logoCharWidth = len(logoChars[0])
	}
}

func placeLogoOnGrid(grid [][]string, startX, startY int) {
	for row := range logoCharHeight {
		for col := range logoCharWidth {
			y := startY + row
			x := startX + col
			if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[y]) {
				grid[y][x] = obstacleStyle.SetString(logoChars[row][col]).Render()
			}
		}
	}
}

var gridStyle = lipgloss.NewStyle().PaddingRight(2).MarginRight(2).Border(lipgloss.ThickBorder(), false, true, false, false)

var logo = lipgloss.NewStyle().
	Width(sidebarWidth).
	SetString(`
░█░█░█▀█░▀▀█░█▀█░█▀▄░█▀▄
░█▀█░█▀█░▄▀░░█▀█░█▀▄░█░█
░▀░▀░▀░▀░▀▀▀░▀░▀░▀░▀░▀▀░
`).
	Render()

var heading = lipgloss.NewStyle().
	Bold(true).
	Underline(true).
	Width(sidebarWidth)

var sidebarValueStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#8d949c"))
