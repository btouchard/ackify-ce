package handlers

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type checkService interface {
	CheckUserSignature(ctx context.Context, docID, userIdentifier string) (bool, error)
}

// BadgeHandler handles badge generation
type BadgeHandler struct {
	checkService checkService
}

// NewBadgeHandler creates a new badge handler
func NewBadgeHandler(checkService checkService) *BadgeHandler {
	return &BadgeHandler{
		checkService: checkService,
	}
}

// HandleStatusPNG generates a PNG badge showing signature status
func (h *BadgeHandler) HandleStatusPNG(w http.ResponseWriter, r *http.Request) {
	docID, err := validateDocID(r)
	if err != nil {
		HandleError(w, models.ErrInvalidDocument)
		return
	}

	userIdentifier, err := validateUserIdentifier(r)
	if err != nil {
		HandleError(w, models.ErrInvalidUser)
		return
	}

	ctx := r.Context()
	isSigned, err := h.checkService.CheckUserSignature(ctx, docID, userIdentifier)
	if err != nil {
		HandleError(w, err)
		return
	}

	badge := h.generateBadge(isSigned)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(badge)
}

const badgeSize = 64

// BadgeColors represents the color scheme for badges
type BadgeColors struct {
	Background color.RGBA
	Icon       color.RGBA
	Border     color.RGBA
}

// BadgeThemes contains predefined color schemes
var BadgeThemes = struct {
	Success BadgeColors
	Error   BadgeColors
}{
	Success: BadgeColors{
		Background: color.RGBA{R: 240, G: 253, B: 244, A: 255}, // success-50
		Icon:       color.RGBA{R: 34, G: 197, B: 94, A: 255},   // success-500
		Border:     color.RGBA{R: 134, G: 239, B: 172, A: 255}, // success-300
	},
	Error: BadgeColors{
		Background: color.RGBA{R: 254, G: 242, B: 242, A: 255}, // red-50
		Icon:       color.RGBA{R: 239, G: 68, B: 68, A: 255},   // red-500
		Border:     color.RGBA{R: 252, G: 165, B: 165, A: 255}, // red-300
	},
}

// generateBadge creates a PNG badge
func (h *BadgeHandler) generateBadge(isSigned bool) []byte {
	img := image.NewRGBA(image.Rect(0, 0, badgeSize, badgeSize))

	colors := h.getBadgeColors(isSigned)
	h.drawBackground(img, colors.Background)
	h.drawBorder(img, colors.Border)
	h.drawIcon(img, isSigned, colors.Icon)

	return h.encodeToPNG(img)
}

// getBadgeColors returns appropriate colors based on signing status
func (h *BadgeHandler) getBadgeColors(isSigned bool) BadgeColors {
	if isSigned {
		return BadgeThemes.Success
	}
	return BadgeThemes.Error
}

// drawBackground fills the image with background color
func (h *BadgeHandler) drawBackground(img *image.RGBA, bgColor color.RGBA) {
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
}

// drawBorder draws a circular border around the badge
func (h *BadgeHandler) drawBorder(img *image.RGBA, borderColor color.RGBA) {
	cx, cy, r := badgeSize/2, badgeSize/2, badgeSize/2-3
	for y := 0; y < badgeSize; y++ {
		for x := 0; x < badgeSize; x++ {
			dx, dy := x-cx, y-cy
			dist := dx*dx + dy*dy
			if dist >= (r*r) && dist <= ((r+2)*(r+2)) {
				img.Set(x, y, borderColor)
			}
		}
	}
}

// drawIcon draws the appropriate icon based on signing status
func (h *BadgeHandler) drawIcon(img *image.RGBA, isSigned bool, iconColor color.RGBA) {
	if isSigned {
		h.drawCheckmark(img, badgeSize, iconColor)
	} else {
		h.drawX(img, badgeSize, iconColor)
	}
}

// encodeToPNG encodes the image to PNG format
func (h *BadgeHandler) encodeToPNG(img *image.RGBA) []byte {
	buf := bytes.NewBuffer(nil)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

// drawCheckmark draws a checkmark icon
func (h *BadgeHandler) drawCheckmark(img *image.RGBA, size int, col color.RGBA) {
	cx, cy := size/2, size/2
	scale := float64(size) / 64.0

	// Checkmark path points (scaled)
	points := [][2]int{
		{int(18 * scale), int(32 * scale)},
		{int(28 * scale), int(42 * scale)},
		{int(46 * scale), int(22 * scale)},
	}

	thickness := int(3 * scale)
	if thickness < 2 {
		thickness = 2
	}

	// Draw first stroke (left part of check)
	h.drawThickLine(img, cx+points[0][0]-cx, cy+points[0][1]-cy,
		cx+points[1][0]-cx, cy+points[1][1]-cy, thickness, col)

	// Draw second stroke (right part of check)
	h.drawThickLine(img, cx+points[1][0]-cx, cy+points[1][1]-cy,
		cx+points[2][0]-cx, cy+points[2][1]-cy, thickness, col)
}

// drawX draws an X icon
func (h *BadgeHandler) drawX(img *image.RGBA, size int, col color.RGBA) {
	cx, cy := size/2, size/2
	offset := int(float64(size) * 0.3)
	thickness := size / 12
	if thickness < 2 {
		thickness = 2
	}

	// Draw diagonal lines for X
	h.drawThickLine(img, cx-offset, cy-offset, cx+offset, cy+offset, thickness, col)
	h.drawThickLine(img, cx-offset, cy+offset, cx+offset, cy-offset, thickness, col)
}

// drawThickLine draws a thick line using Bresenham's algorithm
func (h *BadgeHandler) drawThickLine(img *image.RGBA, x0, y0, x1, y1, thickness int, col color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	x, y := x0, y0
	for {
		// Draw thick point
		for i := -thickness / 2; i <= thickness/2; i++ {
			for j := -thickness / 2; j <= thickness/2; j++ {
				px, py := x+i, y+j
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, col)
				}
			}
		}

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
