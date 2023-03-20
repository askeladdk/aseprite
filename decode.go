package aseprite

import (
	"image"
	"image/color"
	"io"
)

// Read decodes an Aseprite image from r.
func Read(r io.Reader) (*Aseprite, error) {
	var spr Aseprite
	if err := spr.readFrom(r); err != nil {
		return nil, err
	}

	return &spr, nil
}

// Decode decodes an Aseprite image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	return Read(r)
}

// DecodeConfig returns the color model and dimensions of an Aseprite image
// without decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	var f file

	if _, err := f.ReadFrom(r); err != nil {
		return image.Config{}, err
	}

	fw, fh := factorPowerOfTwo(len(f.frames))
	if f.framew > f.frameh {
		fw, fh = fh, fw
	}

	var colorModel color.Model

	switch f.bpp {
	case 8:
		f.initPalette()
		colorModel = f.palette
	case 16:
		colorModel = color.Gray16Model
	default:
		colorModel = color.RGBAModel
	}

	return image.Config{
		ColorModel: colorModel,
		Width:      f.framew * fw,
		Height:     f.frameh * fh,
	}, nil
}

func init() {
	image.RegisterFormat("aseprite", "????\xE0\xA5", Decode, DecodeConfig)
}
