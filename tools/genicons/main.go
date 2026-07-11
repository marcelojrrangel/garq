package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type RGBA = color.RGBA

var (
	white   = RGBA{255, 255, 255, 255}
	black   = RGBA{0, 0, 0, 180}
	grey    = RGBA{180, 180, 180, 255}
	dGrey   = RGBA{100, 100, 100, 255}
	yellow  = RGBA{255, 200, 0, 255}
	dYellow = RGBA{180, 140, 0, 255}
	blue    = RGBA{70, 150, 230, 255}
	dBlue   = RGBA{40, 100, 180, 255}
	red     = RGBA{220, 60, 60, 255}
	dRed    = RGBA{160, 30, 30, 255}
	teal    = RGBA{40, 180, 160, 255}
	dTeal   = RGBA{20, 130, 110, 255}
	green   = RGBA{60, 200, 80, 255}
	dGreen  = RGBA{30, 140, 50, 255}
	orange  = RGBA{240, 150, 40, 255}
	dOrange = RGBA{180, 100, 20, 255}
	purple  = RGBA{160, 80, 200, 255}
	dPurple = RGBA{110, 40, 150, 255}
	olive   = RGBA{140, 160, 60, 255}
	dOlive  = RGBA{90, 110, 30, 255}
	cyan    = RGBA{60, 200, 220, 255}
	dCyan   = RGBA{30, 150, 170, 255}
)

type iconDef struct {
	name string
	fn   func() *image.RGBA
}

const S = 28

func main() {
	outDir := filepath.Join("assets", "icons")
	os.MkdirAll(outDir, 0755)

	icons := []iconDef{
		{"folder",   drawFolder},
		{"file",     drawFile},
		{"cut",      drawCut},
		{"copy",     drawCopy},
		{"paste",    drawPaste},
		{"rename",   drawRename},
		{"delete",   drawDelete},
		{"compress", drawCompress},
		{"extract",  drawExtract},
		{"sort",     drawSort},
		{"preview",  drawPreview},
		{"back",     drawBack},
		{"forward",  drawForward},
		{"up",       drawUp},
		{"newtab",   drawNewTab},
		{"closetab", drawCloseTab},
	}

	for _, ic := range icons {
		img := ic.fn()
		pngData := encodePNG(img)
		icoData := encodeICO(pngData, S)
		path := filepath.Join(outDir, ic.name+".ico")
		if err := os.WriteFile(path, icoData, 0644); err != nil {
			panic(err)
		}
		println("Gerado:", path)
	}
}

func encodePNG(img *image.RGBA) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func encodeICO(pngData []byte, size int) []byte {
	// ICO header: reserved(2) + type(2) + count(2)
	hdr := make([]byte, 6)
	binary.LittleEndian.PutUint16(hdr[0:2], 0)    // reserved
	binary.LittleEndian.PutUint16(hdr[2:4], 1)    // type = icon
	binary.LittleEndian.PutUint16(hdr[4:6], 1)    // count = 1

	// Directory entry: w(1) + h(1) + palette(1) + reserved(1) + planes(2) + bpp(2) + size(4) + offset(4)
	entry := make([]byte, 16)
	if size >= 256 {
		entry[0] = 0
		entry[1] = 0
	} else {
		entry[0] = byte(size)
		entry[1] = byte(size)
	}
	entry[2] = 0    // no palette
	entry[3] = 0    // reserved
	binary.LittleEndian.PutUint16(entry[4:6], 1)   // planes
	binary.LittleEndian.PutUint16(entry[6:8], 32)  // bpp
	binary.LittleEndian.PutUint32(entry[8:12], uint32(len(pngData))) // size
	binary.LittleEndian.PutUint32(entry[12:16], 22) // offset = header(6) + entry(16)

	result := append(hdr, entry...)
	result = append(result, pngData...)
	return result
}

func newCanvas(bg RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, S, S))
	for y := 0; y < S; y++ {
		for x := 0; x < S; x++ {
			img.Set(x, y, bg)
		}
	}
	return img
}

func fillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if x >= 0 && x < S && y >= 0 && y < S {
				img.Set(x, y, c)
			}
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.Color) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			if x >= 0 && x < S && y >= 0 && y < S {
				dx, dy := x-cx, y-cy
				if dx*dx+dy*dy <= r*r {
					img.Set(x, y, c)
				}
			}
		}
	}
}

func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, c color.Color) {
	minX := max(0, min(x1, x2, x3))
	maxX := min(S-1, max(x1, x2, x3))
	minY := max(0, min(y1, y2, y3))
	maxY := min(S-1, max(y1, y2, y3))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInTriangle(x, y, x1, y1, x2, y2, x3, y3) {
				img.Set(x, y, c)
			}
		}
	}
}

func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	d1 := sign(px, py, x1, y1, x2, y2)
	d2 := sign(px, py, x2, y2, x3, y3)
	d3 := sign(px, py, x3, y3, x1, y1)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func sign(px, py, x1, y1, x2, y2 int) int {
	return (px-x2)*(y1-y2) - (x1-x2)*(py-y2)
}

func min(a ...int) int {
	m := a[0]
	for _, v := range a[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func max(a ...int) int {
	m := a[0]
	for _, v := range a[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		steps = 1
	}
	for i := 0; i <= steps; i++ {
		x := x1 + dx*i/steps
		y := y1 + dy*i/steps
		if x >= 0 && x < S && y >= 0 && y < S {
			img.Set(x, y, c)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawFolder() *image.RGBA {
	img := newCanvas(RGBA{250, 250, 245, 255})
	fillRect(img, 3, 8, 21, 22, yellow)
	fillRect(img, 3, 8, 21, 22, dYellow)
	fillRect(img, 3, 6, 11, 8, yellow)
	fillRect(img, 3, 6, 11, 8, dYellow)
	fillRect(img, 3, 8, 21, 9, RGBA{255, 230, 100, 255})
	fillRect(img, 3, 21, 21, 22, RGBA{140, 100, 0, 255})
	return img
}

func drawFile() *image.RGBA {
	img := newCanvas(white)
	fillRect(img, 4, 3, 20, 22, blue)
	fillTriangle(img, 16, 3, 20, 3, 20, 7, white)
	fillRect(img, 7, 9, 18, 10, white)
	fillRect(img, 7, 12, 18, 13, white)
	fillRect(img, 7, 15, 18, 16, white)
	fillRect(img, 7, 18, 14, 19, white)
	fillRect(img, 4, 3, 5, 22, dBlue)
	fillRect(img, 19, 8, 20, 22, dBlue)
	fillRect(img, 4, 3, 20, 4, dBlue)
	fillRect(img, 4, 22, 20, 22, dBlue)
	return img
}

func drawCut() *image.RGBA {
	img := newCanvas(white)
	drawLine(img, 19, 3, 8, 14, red)
	drawLine(img, 19, 4, 9, 14, red)
	drawLine(img, 18, 3, 8, 13, dRed)
	drawLine(img, 5, 3, 16, 14, red)
	drawLine(img, 5, 4, 15, 14, red)
	drawLine(img, 6, 3, 16, 13, dRed)
	fillCircle(img, 12, 12, 2, dRed)
	fillCircle(img, 5, 19, 3, red)
	fillCircle(img, 5, 19, 2, dRed)
	fillCircle(img, 19, 19, 3, red)
	fillCircle(img, 19, 19, 2, dRed)
	return img
}

func drawCopy() *image.RGBA {
	img := newCanvas(white)
	fillRect(img, 8, 3, 22, 21, grey)
	fillTriangle(img, 15, 3, 22, 3, 22, 10, white)
	fillRect(img, 11, 6, 20, 7, white)
	fillRect(img, 11, 9, 20, 10, white)
	fillRect(img, 2, 5, 16, 23, blue)
	fillTriangle(img, 12, 5, 16, 5, 16, 9, white)
	fillRect(img, 5, 10, 14, 11, white)
	fillRect(img, 5, 13, 14, 14, white)
	fillRect(img, 5, 16, 14, 17, white)
	fillRect(img, 5, 19, 10, 20, white)
	fillRect(img, 2, 5, 3, 23, dBlue)
	fillRect(img, 2, 5, 16, 6, dBlue)
	fillRect(img, 2, 23, 16, 23, dBlue)
	return img
}

func drawPaste() *image.RGBA {
	img := newCanvas(white)
	fillRect(img, 5, 6, 19, 23, green)
	fillRect(img, 7, 3, 17, 8, dGreen)
	fillRect(img, 7, 10, 18, 11, white)
	fillRect(img, 7, 13, 18, 14, white)
	fillRect(img, 7, 16, 18, 17, white)
	fillRect(img, 7, 19, 14, 20, white)
	fillRect(img, 5, 6, 6, 23, dGreen)
	fillRect(img, 5, 6, 19, 7, dGreen)
	fillRect(img, 5, 23, 19, 23, dGreen)
	return img
}

func drawRename() *image.RGBA {
	img := newCanvas(white)
	for i := 0; i < 3; i++ {
		drawLine(img, 6+i, 4, 19-i, 20, orange)
	}
	fillTriangle(img, 5, 22, 8, 18, 3, 23, dOrange)
	fillTriangle(img, 4, 22, 7, 18, 3, 23, orange)
	fillRect(img, 6, 3, 15, 4, dOrange)
	drawLine(img, 6, 20, 7, 19, RGBA{80, 80, 80, 255})
	return img
}

func drawDelete() *image.RGBA {
	img := newCanvas(white)
	fillRect(img, 4, 4, 20, 7, dRed)
	fillRect(img, 5, 7, 19, 22, red)
	fillRect(img, 8, 9, 9, 20, white)
	fillRect(img, 12, 9, 13, 20, white)
	fillRect(img, 16, 9, 17, 20, white)
	fillRect(img, 5, 7, 6, 22, dRed)
	fillRect(img, 18, 7, 19, 22, dRed)
	fillRect(img, 5, 22, 19, 22, dRed)
	fillRect(img, 8, 3, 16, 4, dRed)
	return img
}

func drawCompress() *image.RGBA {
	img := newCanvas(white)
	fillRect(img, 2, 7, 22, 22, purple)
	fillRect(img, 2, 5, 11, 7, purple)
	drawLine(img, 12, 7, 12, 22, white)
	for y := 8; y <= 21; y += 3 {
		fillRect(img, 10, y, 14, y+1, white)
	}
	fillRect(img, 10, 18, 14, 20, dPurple)
	fillRect(img, 2, 7, 3, 22, dPurple)
	fillRect(img, 21, 7, 22, 22, dPurple)
	fillRect(img, 2, 21, 22, 22, dPurple)
	return img
}

func drawExtract() *image.RGBA {
	img := newCanvas(white)
	fillRect(img, 2, 7, 22, 22, cyan)
	fillRect(img, 2, 5, 11, 7, cyan)
	fillRect(img, 16, 12, 23, 16, dCyan)
	fillTriangle(img, 23, 10, 23, 18, 19, 14, dCyan)
	fillRect(img, 2, 7, 3, 22, dCyan)
	fillRect(img, 2, 21, 22, 22, dCyan)
	return img
}

func drawSort() *image.RGBA {
	img := newCanvas(white)
	fillTriangle(img, 12, 2, 5, 11, 19, 11, green)
	fillRect(img, 10, 11, 14, 14, green)
	fillTriangle(img, 12, 23, 5, 14, 19, 14, dGrey)
	fillRect(img, 10, 11, 14, 14, dGrey)
	return img
}

func drawPreview() *image.RGBA {
	img := newCanvas(white)
	fillCircle(img, 12, 12, 11, RGBA{220, 230, 240, 255})
	fillCircle(img, 12, 12, 8, RGBA{100, 170, 220, 255})
	fillCircle(img, 12, 12, 7, RGBA{60, 140, 200, 255})
	fillCircle(img, 12, 12, 4, RGBA{30, 60, 100, 255})
	fillCircle(img, 10, 9, 2, white)
	for a := -140; a <= -40; a++ {
		rad := float64(a) * math.Pi / 180
		x := 12 + int(5*math.Cos(rad))
		y := 2 + int(5*math.Sin(rad))
		if x >= 0 && x < S && y >= 0 && y < S {
			img.Set(x, y, dGrey)
		}
	}
	return img
}

func drawBack() *image.RGBA {
	img := newCanvas(white)
	fillTriangle(img, 6, 14, 21, 5, 21, 23, blue)
	drawLine(img, 6, 14, 21, 5, dBlue)
	drawLine(img, 6, 14, 21, 23, dBlue)
	drawLine(img, 21, 5, 21, 23, dBlue)
	return img
}

func drawForward() *image.RGBA {
	img := newCanvas(white)
	fillTriangle(img, 22, 14, 7, 5, 7, 23, blue)
	drawLine(img, 22, 14, 7, 5, dBlue)
	drawLine(img, 22, 14, 7, 23, dBlue)
	drawLine(img, 7, 5, 7, 23, dBlue)
	return img
}

func drawUp() *image.RGBA {
	img := newCanvas(white)
	fillTriangle(img, 14, 5, 6, 20, 22, 20, blue)
	drawLine(img, 14, 5, 6, 20, dBlue)
	drawLine(img, 14, 5, 22, 20, dBlue)
	drawLine(img, 6, 20, 22, 20, dBlue)
	return img
}

func drawNewTab() *image.RGBA {
	img := newCanvas(white)
	// Horizontal bar of +
	fillRect(img, 6, 12, 22, 16, green)
	fillRect(img, 6, 12, 22, 16, dGreen)
	// Vertical bar of +
	fillRect(img, 12, 6, 16, 22, green)
	fillRect(img, 12, 6, 16, 22, dGreen)
	// Center highlight
	fillRect(img, 13, 13, 15, 15, RGBA{120, 255, 140, 255})
	return img
}

func drawCloseTab() *image.RGBA {
	img := newCanvas(white)
	drawLine(img, 7, 7, 21, 21, red)
	drawLine(img, 7, 8, 20, 21, red)
	drawLine(img, 8, 7, 21, 20, red)
	drawLine(img, 7, 21, 21, 7, red)
	drawLine(img, 7, 20, 20, 7, red)
	drawLine(img, 8, 21, 21, 8, red)
	return img
}
