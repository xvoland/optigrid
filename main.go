/*
Copyright © 2026, Vitalii Tereshchuk | DOTOCA.NET All rights reserved.
Homepage: https://dotoca.net/grid-optical-illusion-app

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// --- CONSTANTS ---
const (
	GridSub   = 4 // Visual sub-divisions
	AnimSpeed = 0.08
	SlideTime = 5.0
	SsaaScale = 2.0
	CellSize  = 40 // Square cell size in pixels
	MinGrid   = 8
	MaxGrid   = 50
)

// --- PALETTE ---
var (
	ColorBgBlue   = color.RGBA{63, 81, 151, 255}
	ColorBgLight  = color.RGBA{238, 234, 214, 255}
	ColorGridGray = color.RGBA{160, 160, 160, 255}
	BallDark      = color.RGBA{26, 31, 74, 255}
	BallWhite     = color.RGBA{220, 225, 235, 255}
	BallGlare     = color.RGBA{255, 255, 255, 255}
	BallOutline   = color.RGBA{20, 20, 40, 255}
)

// Ball defines properties for a single animated dot
type Ball struct {
	x, y, tx, ty float32
	color        color.RGBA
	alpha        float32
}

// Update interpolates position and transparency
func (b *Ball) Update(tx, ty float32, visible bool) {
	b.tx, b.ty = tx, ty
	b.x += (b.tx - b.x) * AnimSpeed
	b.y += (b.ty - b.y) * AnimSpeed

	targetAlpha := float32(0)
	if visible {
		targetAlpha = 255
	}
	b.alpha += (targetAlpha - b.alpha) * AnimSpeed
}

// Cell manages logic for a single grid tile
type Cell struct {
	state int
	balls [4]*Ball
}

// Game maintains the core application state
type Game struct {
	cols           int
	rows           int
	cells          [][]*Cell
	autoMode       bool
	lastSwitch     time.Time
	currentIdx     int
	currentSlotPos int
	availableSlots []int
	offscreen      *ebiten.Image
	statusText     string
	statusTimer    time.Time
}

// getExeDir returns the absolute path to the directory where the binary is located
func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// NewGame initializes the grid and default settings
func NewGame() *Game {
	g := &Game{
		cols:       9,
		rows:       21,
		lastSwitch: time.Now(),
		statusText: "Ready. Slots: 1-0 | Resize: [ / ] , - / + | Auto: A",
	}
	g.resizeGrid()
	ebiten.SetWindowSize(g.cols*CellSize, g.rows*CellSize)
	return g
}

func (g *Game) resizeGrid() {
	g.cells = make([][]*Cell, g.rows)
	for r := 0; r < g.rows; r++ {
		g.cells[r] = make([]*Cell, g.cols)
		for c := 0; c < g.cols; c++ {
			isBlue := (r+c)%2 == 0
			ballCol := BallDark
			if isBlue {
				ballCol = BallWhite
			}
			g.cells[r][c] = &Cell{}
			for i := 0; i < 4; i++ {
				g.cells[r][c].balls[i] = &Ball{color: ballCol}
			}
		}
	}
	g.setStatus(fmt.Sprintf("Grid: %dx%d", g.cols, g.rows))
}

func (g *Game) setStatus(msg string) {
	g.statusText = msg
	g.statusTimer = time.Now().Add(3 * time.Second)
}

func (g *Game) scanAvailableSlots() {
	dir := getExeDir()
	pattern := filepath.Join(dir, "slot_*.json")
	files, _ := filepath.Glob(pattern)
	g.availableSlots = nil
	for _, f := range files {
		name := filepath.Base(f)
		var num int
		fmt.Sscanf(name, "slot_%d.json", &num)
		g.availableSlots = append(g.availableSlots, num)
	}
	sort.Ints(g.availableSlots)
}

// Update handles input, physics, and pattern switching
func (g *Game) Update() error {
	cw, ch := float32(CellSize), float32(CellSize)

	// 1. Handle Mouse Clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		c, r := mx/int(cw), my/int(ch)
		if c >= 0 && c < g.cols && r >= 0 && r < g.rows {
			g.cells[r][c].state = (g.cells[r][c].state + 1) % 8
			g.autoMode = false
		}
	}

	// 2. Handle Global Keyboard Actions
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.autoMode = !g.autoMode
		if g.autoMode {
			g.scanAvailableSlots()
			g.currentSlotPos = -1
		}
		g.setStatus(fmt.Sprintf("Auto Mode: %v", g.autoMode))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		for r := 0; r < g.rows; r++ {
			for c := 0; c < g.cols; c++ {
				g.cells[r][c].state = 0
			}
		}
		g.setStatus("Board Cleared")
	}

	// Grid resize: horizontal
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		if g.cols > MinGrid {
			g.cols--
			g.resizeGrid()
			ebiten.SetWindowSize(g.cols*CellSize, g.rows*CellSize)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		if g.cols < MaxGrid {
			g.cols++
			g.resizeGrid()
			ebiten.SetWindowSize(g.cols*CellSize, g.rows*CellSize)
		}
	}

	// Grid resize: vertical
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		if g.rows > MinGrid {
			g.rows--
			g.resizeGrid()
			ebiten.SetWindowSize(g.cols*CellSize, g.rows*CellSize)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		if g.rows < MaxGrid {
			g.rows++
			g.resizeGrid()
			ebiten.SetWindowSize(g.cols*CellSize, g.rows*CellSize)
		}
	}

	// 3. Pattern Saving and Loading (1-9, 0)
	for k := ebiten.Key1; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			slotIdx := int(k - ebiten.Key1 + 1)
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				g.saveSlot(slotIdx)
			} else {
				g.loadSlot(slotIdx)
				g.autoMode = false
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			g.saveSlot(10)
		} else {
			g.loadSlot(10)
			g.autoMode = false
		}
	}

	// 4. Auto-mode playlist cycling
	if g.autoMode && time.Since(g.lastSwitch).Seconds() > SlideTime && len(g.availableSlots) > 0 {
		g.currentSlotPos = (g.currentSlotPos + 1) % len(g.availableSlots)
		g.currentIdx = g.availableSlots[g.currentSlotPos]
		g.loadSlot(g.currentIdx)
		g.lastSwitch = time.Now()
	}

	// 5. Physics: Update all balls
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			cell := g.cells[r][c]
			bx, by := float32(c)*cw, float32(r)*ch
			midX, midY := bx+cw/2, by+ch/2
			sw, sh := cw/GridSub, ch/GridSub
			pts := [4][2]float32{
				{bx + sw/2, by + sh/2}, {bx + cw - sw/2, by + ch - sh/2},
				{bx + cw - sw/2, by + sh/2}, {bx + sw/2, by + ch - sh/2},
			}
			for i, b := range cell.balls {
				vis := (cell.state == 1 && i < 2) ||
					(cell.state == 2 && i >= 2) ||
					(cell.state == 3 && (i == 1 || i == 2)) || // справа
					(cell.state == 4 && (i == 0 || i == 3)) || // слева
					(cell.state == 5 && (i == 0 || i == 2)) || // сверху
					(cell.state == 6 && (i == 1 || i == 3)) || // снизу
					(cell.state == 7)
				targetX, targetY := midX, midY
				if vis {
					targetX, targetY = pts[i][0], pts[i][1]
				}
				b.Update(targetX, targetY, vis)
			}
		}
	}
	return nil
}

// Draw renders the frame
func (g *Game) Draw(screen *ebiten.Image) {
	winW := g.cols * CellSize
	winH := g.rows * CellSize

	// Offscreen buffer for SSAA
	if g.offscreen == nil || g.offscreen.Bounds().Dx() != winW*SsaaScale || g.offscreen.Bounds().Dy() != winH*SsaaScale {
		g.offscreen = ebiten.NewImage(winW*SsaaScale, winH*SsaaScale)
	}
	canvas := g.offscreen
	canvas.Fill(ColorBgLight)
	s := float32(SsaaScale)

	cw, ch := float32(CellSize), float32(CellSize)

	// 1. Draw Checkerboard
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			if (r+c)%2 == 0 {
				vector.DrawFilledRect(canvas, float32(c)*cw*s, float32(r)*ch*s, cw*s, ch*s, ColorBgBlue, true)
			}
		}
	}

	// 2. Draw Grid
	sw, sh := cw/GridSub, ch/GridSub
	for i := 0; i <= g.cols*GridSub; i++ {
		vector.StrokeLine(canvas, float32(i)*sw*s, 0, float32(i)*sw*s, float32(winH)*s, 1*s, ColorGridGray, true)
	}
	for i := 0; i <= g.rows*GridSub; i++ {
		vector.StrokeLine(canvas, 0, float32(i)*sh*s, float32(winW)*s, float32(i)*sh*s, 1*s, ColorGridGray, true)
	}

	// 3. Draw Balls
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			for _, b := range g.cells[r][c].balls {
				if b.alpha < 5 {
					continue
				}
				bc := b.color
				bc.A = uint8(b.alpha)
				px, py, radius := b.x*s, b.y*s, float32(6*s)
				vector.DrawFilledCircle(canvas, px, py, radius, bc, true)
				vector.StrokeCircle(canvas, px, py, radius, 1*s, BallOutline, true)
				vector.DrawFilledCircle(canvas, px-radius/3, py-radius/3, radius/4, BallGlare, true)
			}
		}
	}

	// 4. Blit to Screen with Resize Support
	op := &ebiten.DrawImageOptions{}
	w, h := screen.Size()
	sx := float64(w) / float64(winW*SsaaScale)
	sy := float64(h) / float64(winH*SsaaScale)
	op.GeoM.Scale(sx, sy)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(canvas, op)

	// 5. Draw UI Text (Status Bar)
	if time.Now().Before(g.statusTimer) || g.autoMode {
		msg := g.statusText
		if g.autoMode {
			msg = fmt.Sprintf("[AUTO] Slot: %d", g.currentIdx)
		}
		ebitenutil.DebugPrintAt(screen, msg, 10, h-25)
	}
}

func (g *Game) saveSlot(idx int) {
	data := make([][]int, g.rows)
	for r := 0; r < g.rows; r++ {
		data[r] = make([]int, g.cols)
		for c := 0; c < g.cols; c++ {
			data[r][c] = g.cells[r][c].state
		}
	}
	bytes, _ := json.Marshal(data)
	path := filepath.Join(getExeDir(), fmt.Sprintf("slot_%d.json", idx))
	err := os.WriteFile(path, bytes, 0644)
	if err == nil {
		g.setStatus(fmt.Sprintf("Saved Slot %d", idx))
	} else {
		g.setStatus("Save Error!")
	}
}

func (g *Game) loadSlot(idx int) {
	path := filepath.Join(getExeDir(), fmt.Sprintf("slot_%d.json", idx))
	bytes, err := os.ReadFile(path)
	if err != nil {
		g.setStatus(fmt.Sprintf("Slot %d: File Not Found", idx))
		return
	}
	var data [][]int
	if err := json.Unmarshal(bytes, &data); err == nil {
		for r := 0; r < g.rows && r < len(data); r++ {
			for c := 0; c < g.cols && c < len(data[r]); c++ {
				g.cells[r][c].state = data[r][c]
			}
		}
		g.setStatus(fmt.Sprintf("Loaded Slot %d", idx))
	}
}

// Layout handles scaling logic for resizing windows
func (g *Game) Layout(w, h int) (int, int) {
	return g.cols * CellSize, g.rows * CellSize
}

func main() {
	ebiten.SetWindowTitle("'Grid' Optical Illusion - Pattern Studio")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
