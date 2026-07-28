package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) renderSidebar() (string, *tea.Cursor) {
	inputView, cursor := m.inputs.View()

	sidebar := lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		m.renderCitizenStatus(),
		inputView,
		m.renderHelpKey(),
	)

	return sidebar, cursor
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

func (m Model) renderHelpKey() string {
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

	b.WriteString("\n\n")
	b.WriteString(heading.SetString("Controls").Render())

	b.WriteString("\nPause/Resume: ")
	b.WriteString(helpBarKey.SetString("space").Render())

	b.WriteString("\nReset: ")
	b.WriteString(helpBarKey.SetString("r").Render())

	b.WriteString("\nToggle paths: ")
	b.WriteString(helpBarKey.SetString("p").Render())

	b.WriteString("\nNavigate inputs: ")
	b.WriteString(helpBarKey.SetString("tab").Render())

	b.WriteString("\nQuit: ")
	b.WriteString(helpBarKey.SetString("q").Render())

	return b.String()
}
