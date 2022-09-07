// Copyright (c) 2012 Guillermo Estrada. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
// http://github.com/phrozen/blend

package blend

import (
	"image/color"
	"math"
)

func clampToByte(a float64) byte {
	if a < 0 {
		return 0
	} else if a > math.MaxUint8 {
		return math.MaxUint8
	}
	return byte(a)
}

type rgbaf64 struct {
	r, g, b, a float64
}

func color2rgbaf64(c color.Color) rgbaf64 {
	r, g, b, a := c.RGBA()
	return rgbaf64{float64(r), float64(g), float64(b), float64(a)}
}

type hslf64 struct {
	h, s, l float64
}

func rgb2hsl(c color.Color) hslf64 {
	var h, s, l float64
	cr, cg, cb, _ := c.RGBA()
	r := float64(cr) / max
	g := float64(cg) / max
	b := float64(cb) / max
	cmax := math.Max(math.Max(r, g), b)
	cmin := math.Min(math.Min(r, g), b)
	l = (cmax + cmin) / 2.0
	if cmax != cmin { // Chromatic, else Achromatic.
		delta := cmax - cmin
		if l > 0.5 {
			s = delta / (2.0 - cmax - cmin)
		} else {
			s = delta / (cmax + cmin)
		}
		switch cmax {
		case r:
			h = (g - b) / delta
			if g < b {
				h += 6.0
			}
		case g:
			h = (b-r)/delta + 2.0
		case b:
			h = (r-g)/delta + 4.0
		}
		h /= 6.0
	}
	return hslf64{h, s, l}
}

func hsl2rgb(h, s, l float64) color.Color {
	var r, g, b float64
	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - s*l
		}
		p := 2*l - q
		r = hue2rgb(p, q, h+1.0/3)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3)
	}
	return color.RGBA{
		R: clampToByte(r*math.MaxUint8 + 0.5),
		G: clampToByte(g*math.MaxUint8 + 0.5),
		B: clampToByte(b*math.MaxUint8 + 0.5),
		A: math.MaxUint8,
	}
}

func hue2rgb(p, q, t float64) float64 {
	if t < 0.0 {
		t += 1.0
	} else if t > 1.0 {
		t -= 1.0
	}

	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6.0*t
	case t < 0.5:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6.0
	default:
		return p
	}
}
