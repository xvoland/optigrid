# OptiGrid App - Grid Optical Illusion Pattern Studio

A desktop application for creating and managing ball pattern layouts on a grid.

Check the home page for a more detailed description URL: https://dotoca.net/optigrid


## Features

- **Interactive Grid**: Dynamic grid (8-50 rows/columns), checkerboard pattern, square cells
- **8 Ball States per Cell**:
  - 0: Empty
  - 1: Diagonal ↘ (top-left, bottom-right)
  - 2: Diagonal ↗ (top-right, bottom-left)
  - 3: Right side (both right balls)
  - 4: Left side (both left balls)
  - 5: Top side (both top balls)
  - 6: Bottom side (both bottom balls)
  - 7: All 4 balls

- **10 Save Slots**: Store and load patterns (keys 1-9, 0 for slot 10)
- **Auto Mode**: Cycle through slots automatically
- **Smooth Animation**: Ball movement interpolation
- **Window Resizing**: Full window resize support
- **Optical Illusion Mode**: Create mesmerizing patterns

## Controls

| Key | Action |
|-----|--------|
| Left Click | Cycle cell state (0→1→2→3→4→5→6→7→0) |
| 1-9, 0 | Load slot 1-9, 10 |
| Shift+1-9, Shift+0 | Save to slot |
| A | Toggle Auto Mode |
| C | Clear all cells |
| [ | Decrease columns (min 8) |
| ] | Increase columns (max 50) |
| - | Decrease rows (min 8) |
| + | Increase rows (max 50) |

## Optical Illusion Mode

This application is inspired by the famous "Grid" optical illusion where small circles inside a grid create virtual concentric rings that trick the brain into seeing a bubble distortion despite all lines being perfectly parallel.

With 8 ball states per cell (including state 7 with all 4 balls), you can create similar mesmerizing patterns that play with visual perception.

**Tips to enhance the effect:**
- Use state 7 (all 4 balls) for maximum visual impact
- Create patterns that guide the eye toward the center
- Try the Auto Mode to cycle through saved patterns

## Build

```bash
go mod init pattern-studio
go mod tidy
go build -o pattern-studio
```

## Requirements

- Go 1.20+
- Ebiten v2

## License

MIT