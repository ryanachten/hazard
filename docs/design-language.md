# Visual Design Language

A monospace character set + colour system for rendering the simulation grid.

## Philosophy

- Each **type** (obstacle, fire, flood, etc.) has its own set of characters
- Each **instance** (a specific building, hazard zone, safe zone) picks one character from its type palette and uses it uniformly across all its cells
- Citizens overlay whatever tile they stand on — you always see the agent first
- Colour reinforces type at a glance; character adds texture
- Open ground varies subtly cell-to-cell for background texture

## Core Tiles

| Type | Glyph palette | Colour | Notes |
|---|---|---|---|
| Open ground | `·` `.` `·` `·` `.` `·` `·` ` ` | `#586069` dim gray | Subtle cell-to-cell variation; never obscures |
| Obstacle | `#` `▓` `▒` `≡` | `#6b7280` slate | Solid, blocked — building/terrain |
| Safe zone | `◎` `◉` `◌` | `#22c55e` green | Goal area; one char per zone |
| Citizen (navigating) | `@` | `#eab308` yellow | Classic roguelike actor symbol |
| Citizen (escaped) | `♦` | `#4ade80` bright green | Reached safety |
| Citizen (dead) | `†` | `#991b1b` dark red | Overtaken by hazard |

## Hazard Palettes (type-specific)

| Type | Glyph palette | Colour | Rationale |
|---|---|---|---|
| Fire | `░` `"` `^` | `#ef4444` red | Flickering flames, sparks, embers |
| Flood | `~` `≈` `░` | `#3b82f6` blue | Waves, ripples — fluid motion |
| Lava | `≈` `▒` `°` `o` | `#f97316` orange | Flowing streams, glowing crust, gas bubbles |

Each hazard zone instance picks one character from its type palette and uses it uniformly.

## Per-Cell Background Colours

Every tile has an explicit background colour — no transparency. This keeps the
grid consistent across terminal themes.

| Type | Background | Rationale |
|---|---|---|
| Open ground | `#0d1117` | Matches the terminal chrome; invisible seam between grid and chrome |
| Obstacle | `#1c1f2b` | Subtle dark slate — reads as solid mass |
| Fire | `#2a0a0a` | Deep ember glow — hints at heat without washing out red glyphs |
| Flood | `#0a1628` | Dark water tint — cooling counterpoint to fire |
| Lava | `#2a1400` | Warm dark orange — background tells you it's hot before you read the glyph |
| Safe zone | `#0d2a0d` | Dark green — frames the goal area without competing with bright `◎` |
| Citizen (navigating) | `#2a2000` | Subtle yellow highlight — helps spot agents at a glance |
| Citizen (escaped) | `#0d2a0d` | Matches safe zone — reached safety |
| Citizen (dead) | `#2a0505` | Deep blood red — clearly signals loss |

In Bubble Tea, set these via `lipgloss.NewStyle().Background(lipgloss.Color("..."))`.
Each cell is rendered as a styled string; no window-wide background is needed.

## Layering Order (bottom to top)

1. Open ground (background)
2. Obstacle / hazard / safe zone tiles (overwrite ground)
3. Citizens (overwrite tile beneath them)

## Quick Reference

```text
·  open ground         ◎  safe zone          @  citizen (alive)
#  obstacle (bldg)     ~  flood              ♦  citizen (escaped)
░  fire                ≈  lava               †  citizen (dead)
```


