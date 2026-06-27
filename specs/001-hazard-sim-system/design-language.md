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

## Layering Order (bottom to top)

1. Open ground (background)
2. Obstacle / hazard / safe zone tiles (overwrite ground)
3. Citizens (overwrite tile beneath them)

## Quick Reference

```
·  open ground         ◎  safe zone          @  citizen (alive)
#  obstacle (bldg)     ~  flood              ♦  citizen (escaped)
░  fire                ≈  lava               †  citizen (dead)
```

See `design-demo.html` for an interactive reference of glyphs and colors.
