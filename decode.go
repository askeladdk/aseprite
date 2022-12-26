package aseprite

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

// Decode reads a Aseprite image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	var spr Aseprite
	if err := spr.readFrom(r); err != nil {
		return nil, err
	}

	return &spr, nil
}

// DecodeConfig returns the color model and dimensions of an Aseprite image
// without decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	var raw [14]byte

	if _, err := io.ReadFull(r, raw[:]); err != nil {
		return image.Config{}, err
	}

	if magic := binary.LittleEndian.Uint16(raw[4:]); magic != 0xA5E0 {
		return image.Config{}, errInvalidMagic
	}

	nframes := int(binary.LittleEndian.Uint16(raw[6:]))
	framew := int(binary.LittleEndian.Uint16(raw[8:]))
	frameh := int(binary.LittleEndian.Uint16(raw[10:]))
	bpp := binary.LittleEndian.Uint16(raw[12:])

	colorModel := color.RGBAModel
	if bpp == 16 {
		colorModel = color.Gray16Model
	}

	fw, fh := factorPowerOfTwo(nframes)
	if framew > frameh {
		fw, fh = fh, fw
	}

	return image.Config{
		ColorModel: colorModel,
		Width:      framew * fw,
		Height:     frameh * fh,
	}, nil
}

func init() {
	image.RegisterFormat("aseprite", "????\xE0\xA5", Decode, DecodeConfig)
}
