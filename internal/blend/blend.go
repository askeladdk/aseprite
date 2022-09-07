// Copyright (c) 2012 Guillermo Estrada. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
// http://github.com/phrozen/blend

package blend

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Constants of max and mid values for uint16 for internal use.
// This can be changed to make the algorithms use uint8 instead,
// but they are kept this way to provide more accurate calculations
// and to support all of the color modes in the 'image' package.
const (
	max = float64(math.MaxUint16) // max range of color.Color
	mid = max / 2
)

// BlendFunc blends the source color with the destination color.
type BlendFunc func(dst, src color.Color) color.Color

// Modes lists all blend modes that are supported by the Aseprite file format.
var Modes = [19]BlendFunc{
	0:  Normal,
	1:  Multiply,
	2:  Screen,
	3:  Overlay,
	4:  Darken,
	5:  Lighten,
	6:  ColorDodge,
	7:  ColorBurn,
	8:  HardLight,
	9:  SoftLight,
	10: Difference,
	11: Exclusion,
	12: Hue,
	13: Saturation,
	14: Color,
	15: Luminosity,
	16: Addition,
	17: Subtract,
	18: Divide,
}

// clip clips r against each image's bounds (after translating into the
// destination image's coordinate space) and shifts the points sp and mp by
// the same amount as the change in r.Min.
func clip(dst image.Image, r *image.Rectangle, src0 image.Image,
	sp0 *image.Point, src1 image.Image, sp1 *image.Point) {
	orig := r.Min
	*r = r.Intersect(dst.Bounds())
	*r = r.Intersect(src0.Bounds().Add(orig.Sub(*sp0)))
	*r = r.Intersect(src1.Bounds().Add(orig.Sub(*sp1)))
	dx := r.Min.X - orig.X
	dy := r.Min.Y - orig.Y
	sp0.X += dx
	sp0.Y += dy
	sp1.X += dx
	sp1.Y += dy
}

// Blend blends src0 (top layer) into src1 (bottom layer) using mode
// and stores the result in dst.
func Blend(dst draw.Image, r image.Rectangle, src0 image.Image, sp0 image.Point,
	src1 image.Image, sp1 image.Point, mode BlendFunc) {
	clip(dst, &r, src0, &sp0, src1, &sp1)
	if r.Empty() {
		return
	}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			src0col := src0.At(x+sp0.X, y+sp0.Y)
			// aseprite does not blend transparent pixels.
			if _, _, _, a := src0col.RGBA(); a == 0 {
				dst.Set(x, y, src0col)
			} else {
				dst.Set(x, y, mode(src1.At(x+sp1.X, y+sp1.Y), src0col))
			}
		}
	}
}

func blendPerChannel(dst, src color.Color, blend func(float64, float64) float64) color.Color {
	d := color2rgbaf64(dst)
	s := color2rgbaf64(src)
	return color.RGBA{
		R: clampToByte(blend(d.r, s.r)/256 + 0.5),
		G: clampToByte(blend(d.g, s.g)/256 + 0.5),
		B: clampToByte(blend(d.b, s.b)/256 + 0.5),
		A: clampToByte(d.a/256 + 0.5),
	}
}

// Normal.
func Normal(dst, src color.Color) color.Color {
	return src
}

// Darken.
func Darken(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return math.Min(d, s)
	})
}

// Multiply.
func Multiply(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return s * d / max
	})
}

// Color burn.
func ColorBurn(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		if s == 0.0 {
			return s
		}
		return math.Max(0.0, max-((max-d)*max/s))
	})
}

// Lighten.
func Lighten(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return math.Max(d, s)
	})
}

// Screen.
func Screen(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return s + d - s*d/max
	})
}

// Color dodge.
func ColorDodge(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		if s == max {
			return s
		}
		return math.Min(max, (d * max / (max - s)))
	})
}

// Overlay.
func Overlay(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		if d < mid {
			return 2 * s * d / max
		}
		return max - 2*(max-s)*(max-d)/max
	})
}

// Soft Light.
func SoftLight(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return (d / max) * (d + (2*s/max)*(max-d))
	})
}

// Hard Light.
func HardLight(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		if s > mid {
			return d + (max-d)*((s-mid)/mid)
		}
		return d * s / mid
	})
}

// Addition.
func Addition(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		if s+d > max {
			return max
		}
		return s + d
	})
}

// Subtract.
func Subtract(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		if d-s < 0.0 {
			return 0.0
		}
		return d - s
	})
}

// Divide.
func Divide(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return (d*max)/s + 1.0
	})
}

// Difference.
func Difference(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return math.Abs(s - d)
	})
}

// Exclusion.
func Exclusion(dst, src color.Color) color.Color {
	return blendPerChannel(dst, src, func(d, s float64) float64 {
		return s + d - s*d/mid
	})
}

// Hue.
func Hue(dst, src color.Color) color.Color {
	s := rgb2hsl(src)
	if s.s == 0.0 {
		return dst
	}
	d := rgb2hsl(dst)
	return hsl2rgb(s.h, d.s, d.l)
}

// Saturation.
func Saturation(dst, src color.Color) color.Color {
	s := rgb2hsl(src)
	d := rgb2hsl(dst)
	return hsl2rgb(d.h, s.s, d.l)
}

// Color.
func Color(dst, src color.Color) color.Color {
	s := rgb2hsl(src)
	d := rgb2hsl(dst)
	return hsl2rgb(s.h, s.s, d.l)
}

// Luminosity.
func Luminosity(dst, src color.Color) color.Color {
	s := rgb2hsl(src)
	d := rgb2hsl(dst)
	return hsl2rgb(d.h, d.s, s.l)
}
