package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mobil-koeln/moko-cli/internal/models"
)

type mapCellType int

const (
	mapCellEmpty mapCellType = iota
	mapCellPath
	mapCellPast
	mapCellCurrent
	mapCellFuture
)

type mapCell struct {
	ch    rune
	ctype mapCellType
}

// renderRouteMap renders a dots-only geographic map of the journey route.
func renderRouteMap(stops []models.Stop, currentIdx, width, height int) string {
	if len(stops) == 0 || width < 3 || height < 3 {
		return ""
	}

	// Filter stops with valid coordinates
	type stopEntry struct {
		index int
		stop  models.Stop
	}
	var valid []stopEntry
	for i, s := range stops {
		if s.Lat != 0 || s.Lon != 0 {
			valid = append(valid, stopEntry{index: i, stop: s})
		}
	}
	if len(valid) == 0 {
		return ""
	}

	// Compute bounding box
	minLat, maxLat := valid[0].stop.Lat, valid[0].stop.Lat
	minLon, maxLon := valid[0].stop.Lon, valid[0].stop.Lon
	for _, v := range valid[1:] {
		if v.stop.Lat < minLat {
			minLat = v.stop.Lat
		}
		if v.stop.Lat > maxLat {
			maxLat = v.stop.Lat
		}
		if v.stop.Lon < minLon {
			minLon = v.stop.Lon
		}
		if v.stop.Lon > maxLon {
			maxLon = v.stop.Lon
		}
	}

	// Handle degenerate cases
	latSpan := maxLat - minLat
	lonSpan := maxLon - minLon
	if latSpan < 0.01 {
		mid := (minLat + maxLat) / 2
		minLat = mid - 0.005
		maxLat = mid + 0.005
		latSpan = 0.01
	}
	if lonSpan < 0.01 {
		mid := (minLon + maxLon) / 2
		minLon = mid - 0.005
		maxLon = mid + 0.005
		lonSpan = 0.01
	}

	// Add 10% padding
	latPad := latSpan * 0.1
	lonPad := lonSpan * 0.1
	minLat -= latPad
	maxLat += latPad
	minLon -= lonPad
	maxLon += lonPad
	latSpan = maxLat - minLat
	lonSpan = maxLon - minLon

	// Scale factors with terminal aspect ratio correction (chars ~2x tall as wide)
	xScale := float64(width-1) / lonSpan
	yScale := float64(height-1) / latSpan * 2.0

	// Use the smaller scale to fit both axes
	scale := xScale
	if yScale < scale {
		scale = yScale
	}

	// Center the map within the available area
	usedWidth := scale * lonSpan
	usedHeight := scale * latSpan / 2.0
	xOffset := (float64(width-1) - usedWidth) / 2
	yOffset := (float64(height-1) - usedHeight) / 2

	// Convert coordinates to grid positions
	type gridPoint struct {
		col int
		row int
	}
	points := make([]gridPoint, len(valid))
	for i, v := range valid {
		col := int(math.Round((v.stop.Lon-minLon)*scale + xOffset))
		row := int(math.Round((maxLat-v.stop.Lat)*scale/2.0 + yOffset))
		if col < 0 {
			col = 0
		}
		if col >= width {
			col = width - 1
		}
		if row < 0 {
			row = 0
		}
		if row >= height {
			row = height - 1
		}
		points[i] = gridPoint{col: col, row: row}
	}

	// Create grid
	grid := make([][]mapCell, height)
	for r := 0; r < height; r++ {
		grid[r] = make([]mapCell, width)
		for c := 0; c < width; c++ {
			grid[r][c] = mapCell{ch: ' ', ctype: mapCellEmpty}
		}
	}

	// Draw route lines between consecutive stops
	for i := 0; i < len(points)-1; i++ {
		bresenhamLine(grid, points[i].col, points[i].row, points[i+1].col, points[i+1].row)
	}

	// Place stop markers
	for i, v := range valid {
		p := points[i]
		var marker rune
		var ct mapCellType
		if v.index < currentIdx {
			marker = '○'
			ct = mapCellPast
		} else if v.index == currentIdx {
			marker = '◉'
			ct = mapCellCurrent
		} else {
			marker = '●'
			ct = mapCellFuture
		}
		grid[p.row][p.col] = mapCell{ch: marker, ctype: ct}
	}

	// Render grid to styled string
	pathStyle := lipgloss.NewStyle().Foreground(colorGray)
	pastStyle := lipgloss.NewStyle().Foreground(colorGray)
	currentStyle := lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	futureStyle := lipgloss.NewStyle().Foreground(colorCyan).Bold(true)

	var b strings.Builder
	for r := 0; r < height; r++ {
		var line strings.Builder
		for c := 0; c < width; c++ {
			ch := string(grid[r][c].ch)
			switch grid[r][c].ctype {
			case mapCellPath:
				line.WriteString(pathStyle.Render(ch))
			case mapCellPast:
				line.WriteString(pastStyle.Render(ch))
			case mapCellCurrent:
				line.WriteString(currentStyle.Render(ch))
			case mapCellFuture:
				line.WriteString(futureStyle.Render(ch))
			default:
				line.WriteString(ch)
			}
		}
		b.WriteString(line.String())
		if r < height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// bresenhamLine draws a line between two points on the grid using Bresenham's algorithm.
func bresenhamLine(grid [][]mapCell, x0, y0, x1, y1 int) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		if y0 >= 0 && y0 < len(grid) && x0 >= 0 && x0 < len(grid[y0]) {
			if grid[y0][x0].ctype == mapCellEmpty {
				grid[y0][x0] = mapCell{ch: '·', ctype: mapCellPath}
			}
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}
