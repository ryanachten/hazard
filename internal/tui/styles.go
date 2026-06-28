package tui

import (
	pf "hazard/internal/pathfinding"
)

var cellCharacters = map[pf.CellType][4]rune{
	pf.CellOpen:     {'·', '.', ',', ' '},
	pf.CellObstacle: {'#', '▓', '▒', '≡'},
	pf.CellHazard:   {'~', '≈', '°', '^'},
	pf.CellSafeZone: {'x', 'x', 'x', 'x'},
}

// TODO: we could add variation or personality here too perhaps
var citizenCharacter = '@'
