package tiff

import (
	"image"
	"image/color"
)

// CMYKA represents CMYKAImg color, having 8 bits for each of cyan,
// magenta, yellow and black, with alpha channel
//
// It is not associated with any particular color profile.
type CMYKA struct {
	C, M, Y, K, A uint8
}

func (c CMYKA) RGBA() (uint32, uint32, uint32, uint32) {
	r, g, b, _ := color.CMYK{
		C: c.C,
		M: c.M,
		Y: c.Y,
		K: c.K,
	}.RGBA()

	w := 0xffff - uint32(c.K)*0x101
	a := uint32((0xffff - uint32(c.A)*0x101) * w / 0xffff)
	return r, g, b, 65535 - a
}

// CMYKAModel is the Model for CMYKAImg colors.
var CMYKAModel color.Model = color.ModelFunc(cmykModel)

func cmykModel(c color.Color) color.Color {
	if _, ok := c.(CMYKA); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	cc, mm, yy, kk := color.RGBToCMYK(uint8(r>>8), uint8(g>>8), uint8(b>>8))
	return CMYKA{cc, mm, yy, kk, 0xff}
}

// CMYKAImg is an in-memory image whose At method returns color.CMYK values.
type CMYKAImg struct {
	// Pix holds the image's pixels, in C, M, Y, K order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

func (p *CMYKAImg) ColorModel() color.Model { return color.CMYKModel }

func (p *CMYKAImg) Bounds() image.Rectangle { return p.Rect }

func (p *CMYKAImg) At(x, y int) color.Color {
	return p.CMYKAt(x, y)
}

func (p *CMYKAImg) RGBA64At(x, y int) color.RGBA64 {
	r, g, b, a := p.CMYKAt(x, y).RGBA()
	return color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)}
}

func (p *CMYKAImg) CMYKAt(x, y int) CMYKA {
	if !(image.Point{x, y}.In(p.Rect)) {
		return CMYKA{}
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+5 : i+5] // Small cap improves performance, see https://golang.org/issue/27857
	return CMYKA{s[0], s[1], s[2], s[3], s[4]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *CMYKAImg) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*5
}

func (p *CMYKAImg) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	c1 := color.CMYKModel.Convert(c).(CMYKA)
	s := p.Pix[i : i+5 : i+5] // Small cap improves performance, see https://golang.org/issue/27857
	s[0] = c1.C
	s[1] = c1.M
	s[2] = c1.Y
	s[3] = c1.K
	s[4] = c1.A
}

func (p *CMYKAImg) SetCMYKA(x, y int, c CMYKA) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+5 : i+5] // Small cap improves performance, see https://golang.org/issue/27857
	s[0] = c.C
	s[1] = c.M
	s[2] = c.Y
	s[3] = c.K
	s[4] = c.A
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *CMYKAImg) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	// If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be inside
	// either r1 or r2 if the intersection is empty. Without explicitly checking for
	// this, the Pix[i:] expression below can panic.
	if r.Empty() {
		return &CMYKAImg{}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &CMYKAImg{
		Pix:    p.Pix[i:],
		Stride: p.Stride,
		Rect:   r,
	}
}

// Opaque scans the entire image and reports whether it is fully opaque.
func (p *CMYKAImg) Opaque() bool {
	return false
}

// NewCMYKA returns a new CMYKAImg image with the given bounds.
func NewCMYKA(r image.Rectangle) *CMYKAImg {
	return &CMYKAImg{
		Pix:    make([]uint8, 5*r.Dx()*r.Dy()),
		Stride: 5 * r.Dx(),
		Rect:   r,
	}
}
