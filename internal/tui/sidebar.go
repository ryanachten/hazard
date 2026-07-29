package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) renderSidebar() (string, *tea.Cursor) {
	var sections []string
	var cursor *tea.Cursor
	used := 0

	sections = append(sections, logo)
	used += lineCount(logo)

	statusStr := m.renderCitizenStatus()
	if used+lineCount(statusStr) <= m.height {
		sections = append(sections, statusStr)
		used += lineCount(statusStr)
	}

	controlsStr := m.renderControls()
	if used+lineCount(controlsStr) <= m.height {
		sections = append(sections, controlsStr)
		used += lineCount(controlsStr)
	}

	keyStr := m.renderKey()
	if used+lineCount(keyStr) <= m.height {
		sections = append(sections, keyStr)
		used += lineCount(keyStr)
	}

	inputView, inputCursor := m.inputs.View()
	if used+lineCount(inputView) <= m.height {
		sections = append(sections, inputView)
		if inputCursor != nil {
			c := *inputCursor
			c.Y += used
			c.X += m.width + 5
			cursor = &c
		}
	}

	sidebar := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return sidebar, cursor
}

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func (m Model) renderCitizenStatus() string {
	activeCitizens := citizenStyle.
		SetString(fmt.Sprintf("Active: %d", m.activeCitizenCount)).
		MarginRight(1).
		Render()

	deadCitizens := citizenDeadStyle.
		SetString(fmt.Sprintf("Dead: %d", m.deadCitizenCount)).
		MarginRight(1).
		Render()

	escapedCitizens := citizenEscapedStyle.
		SetString(fmt.Sprintf("Escaped: %d", m.escapedCitizenCount)).
		MarginBottom(1).
		Render()

	return activeCitizens + deadCitizens + escapedCitizens
}

func (m Model) renderKey() string {
	var b strings.Builder

	b.WriteString(heading.SetString("Key").Render())

	b.WriteString("\nActive citizen: ")
	b.WriteString(citizenCharacter)

	b.WriteString("\nCitizen Path: ")
	b.WriteString(pathHorizontalCharacter)
	b.WriteString(pathVerticalCharacter)

	b.WriteString("\nDead citizen: ")
	b.WriteString(citizenDeadCharacter)

	b.WriteString("\nEscaped citizen: ")
	b.WriteString(citizenEscapedCharacter)

	b.WriteString("\nSafe zone: ")
	for _, char := range safeZoneCharacters {
		b.WriteString(safeZoneStyle.SetString(char).Render())
	}

	b.WriteString("\nOpen cell: ")
	for _, char := range openCharacters {
		b.WriteString(openStyle.SetString(char).Render())
	}

	b.WriteString("\nObstacle: ")
	for _, char := range obstacleCharacters {
		b.WriteString(obstacleStyle.SetString(char).Render())
	}

	b.WriteString("\nFire: ")
	for _, char := range fireCharacters {
		b.WriteString(fireStyle.SetString(char).Render())
	}

	b.WriteString("\nLava: ")
	for _, char := range lavaCharacters {
		b.WriteString(lavaStyle.SetString(char).Render())
	}

	b.WriteString("\nFlood: ")
	for _, char := range floodCharacters {
		b.WriteString(floodStyle.SetString(char).Render())
	}

	b.WriteString("\n")

	return b.String()
}

func (m Model) renderControls() string {
	var b strings.Builder

	b.WriteString(heading.SetString("Controls").Render())

	b.WriteString("\nPause/Resume: ")
	b.WriteString(sidebarValueStyle.SetString("space").Render())

	b.WriteString("\nReset: ")
	b.WriteString(sidebarValueStyle.SetString("r").Render())

	b.WriteString("\nToggle paths: ")
	b.WriteString(sidebarValueStyle.SetString("p").Render())

	b.WriteString("\nNavigate inputs: ")
	b.WriteString(sidebarValueStyle.SetString("tab").Render())

	b.WriteString("\nQuit: ")
	b.WriteString(sidebarValueStyle.SetString("q").Render())

	b.WriteString("\n")

	return b.String()
}
