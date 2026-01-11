package aseprite

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"time"

	"github.com/askeladdk/aseprite/internal/blend"
)

var errInvalidMagic = errors.New("invalid magic number")

type cel struct {
	image image.Image
	mask  image.Uniform
	data  []byte
}

func makeCelImage8(f *file, bounds image.Rectangle, opacity byte, pix []byte) cel {
	return cel{
		image: &image.Paletted{
			Pix:     pix,
			Stride:  bounds.Dx(),
			Rect:    bounds,
			Palette: f.palette,
		},
		mask: image.Uniform{color.Alpha{opacity}},
		data: nil,
	}
}

func makeCelImage16(f *file, bounds image.Rectangle, opacity byte, pix []byte) cel {
	img := image.Gray16{
		Pix:    pix,
		Stride: bounds.Dx() * 2,
		Rect:   bounds,
	}

	mask := image.Uniform{color.Alpha{opacity}}

	return cel{&img, mask, nil}
}

func makeCelImage32(f *file, bounds image.Rectangle, opacity byte, pix []byte) cel {
	img := image.NRGBA{
		Pix:    pix,
		Stride: bounds.Dx() * 4,
		Rect:   bounds,
	}

	mask := image.Uniform{color.Alpha{opacity}}

	return cel{&img, mask, nil}
}

type layer struct {
	flags     uint16
	blendMode uint16
	opacity   byte
	data      []byte
}

func (l *layer) Parse(raw []byte) error {
	if typ := binary.LittleEndian.Uint16(raw[2:]); typ == 2 {
		return errors.New("tilemap layers not supported")
	}
	l.flags = binary.LittleEndian.Uint16(raw)
	l.blendMode = binary.LittleEndian.Uint16(raw[10:])
	l.opacity = raw[12]
	return nil
}

type chunk struct {
	typ int
	raw []byte
}

func (c chunk) Reader() io.Reader {
	return bytes.NewReader(c.raw)
}

func (c *chunk) Read(raw []byte) ([]byte, error) {
	chunkLen := binary.LittleEndian.Uint32(raw)
	c.typ = int(binary.LittleEndian.Uint16(raw[4:]))
	c.raw = raw[6:chunkLen]
	return raw[chunkLen:], nil
}

type frame struct {
	dur    time.Duration
	chunks []chunk
	cels   []cel
}

func (f *frame) Read(raw []byte) ([]byte, error) {
	if magic := binary.LittleEndian.Uint16(raw[4:]); magic != 0xF1FA {
		return nil, errInvalidMagic
	}

	// frameLen := binary.LittleEndian.Uint32(raw[0:])
	oldChunks := binary.LittleEndian.Uint16(raw[6:])
	durationMS := binary.LittleEndian.Uint16(raw[8:])
	newChunks := binary.LittleEndian.Uint32(raw[12:])

	f.dur = time.Millisecond * time.Duration(durationMS)

	nchunks := int(newChunks)
	if nchunks == 0 {
		nchunks = int(oldChunks)
	}

	f.chunks = make([]chunk, nchunks)

	raw = raw[16:]

	for i := 0; i < nchunks; i++ {
		var c chunk
		raw, _ = c.Read(raw)
		f.chunks[i] = c
	}

	return raw, nil
}

type file struct {
	frameWidth              int
	frameHeight             int
	flags                   uint16
	bpp                     uint16
	transparentPaletteIndex uint8
	palette                 color.Palette
	frames                  []frame
	layers                  []layer
	makeCelFunc             func(f *file, bounds image.Rectangle, opacity byte, pix []byte) cel
}

func (f *file) ReadFrom(r io.Reader) (int64, error) {
	var hdr [128]byte

	raw := hdr[:]

	if n, err := io.ReadFull(r, raw); err != nil {
		return int64(n), err
	}

	if magic := binary.LittleEndian.Uint16(raw[4:]); magic != 0xA5E0 {
		return 128, errInvalidMagic
	}

	if pixw, pixh := raw[34], raw[35]; pixw != pixh {
		return 128, errors.New("unsupported pixel ratio")
	}

	f.bpp = binary.LittleEndian.Uint16(raw[12:])
	f.flags = binary.LittleEndian.Uint16(raw[14:])
	f.frames = make([]frame, 0, binary.LittleEndian.Uint16(raw[6:]))
	f.frameWidth = int(binary.LittleEndian.Uint16(raw[8:]))
	f.frameHeight = int(binary.LittleEndian.Uint16(raw[10:]))
	f.palette = make(color.Palette, binary.LittleEndian.Uint16(raw[32:]))
	f.transparentPaletteIndex = raw[28]
	fmt.Println(f.transparentPaletteIndex)

	switch f.bpp {
	case 8:
		f.makeCelFunc = makeCelImage8
	case 16:
		f.makeCelFunc = makeCelImage16
	case 32:
		f.makeCelFunc = makeCelImage32
	default:
		return 0, errors.New("invalid color depth")
	}

	for i := range f.palette {
		f.palette[i] = color.Black
	}
	f.palette[f.transparentPaletteIndex] = color.Transparent

	fileSize := int64(binary.LittleEndian.Uint32(raw))
	raw = make([]byte, fileSize-128)

	if n, err := io.ReadFull(r, raw); err != nil {
		return int64(128 + n), err
	}

	for len(raw) > 0 {
		var fr frame
		var err error
		if raw, err = fr.Read(raw); err != nil {
			return fileSize, err
		}

		f.frames = append(f.frames, fr)
	}

	return fileSize, nil
}

func (f *file) buildAtlas() (atlas draw.Image, framesRectangles []image.Rectangle) {
	var atlasRectangle image.Rectangle
	atlasRectangle, framesRectangles = getBounds(len(f.frames), f.frameWidth, f.frameHeight)

	switch f.bpp {
	case 8:
		atlas = image.NewPaletted(atlasRectangle, f.palette)
	case 16:
		atlas = image.NewGray16(atlasRectangle)
	default:
		atlas = image.NewRGBA(atlasRectangle)
	}

	framebounds := image.Rect(0, 0, f.frameWidth, f.frameHeight)

	dstblend := image.NewRGBA(framebounds)
	dst := image.NewRGBA(framebounds)

	transparent := &image.Uniform{color.Transparent}

	for i, fr := range f.frames {
		draw.Draw(dst, framebounds, transparent, image.Point{}, draw.Src)
		for layer, c := range fr.cels {
			if c.image == nil {
				continue
			}

			srcImage := c.image
			srcBounds := srcImage.Bounds()
			srcBoundsMin := srcBounds.Min

			// Correction to avoid palette index errors if a color has been deleted from the Aseprite palette.
			if imgPaletted, ok := srcImage.(*image.Paletted); ok {
				for i := range imgPaletted.Pix {
					if int(imgPaletted.Pix[i]) >= len(f.palette) {
						// Assign a transparent index if the index is outside the palette range.
						imgPaletted.Pix[i] = f.transparentPaletteIndex
					}
				}
			}

			if mode := f.layers[layer].blendMode; mode > 0 && int(mode) < len(blend.Modes) {
				draw.Draw(dstblend, framebounds, transparent, image.Point{}, draw.Src)
				blend.Blend(dstblend, srcBounds.Sub(srcBoundsMin), srcImage, srcBoundsMin, dst, srcBoundsMin, blend.Modes[mode])
				srcImage = dstblend
				srcBoundsMin = image.Point{}
			}

			draw.DrawMask(dst, srcBounds, srcImage, srcBoundsMin, &c.mask, image.Point{}, draw.Over)
		}

		draw.Draw(atlas, framesRectangles[i], dst, image.Point{}, draw.Src)
	}

	return
}

func (f *file) buildUserData() []byte {
	n := 0

	for _, l := range f.layers {
		if l.flags&1 != 0 {
			n += len(l.data)
		}
	}

	for _, fr := range f.frames {
		for _, c := range fr.cels {
			n += len(c.data)
		}
	}

	return make([]byte, 0, n)
}

func (f *file) buildLayerData(userdata []byte) [][]byte {
	ld := make([][]byte, 0, len(f.layers))
	for _, l := range f.layers {
		if l.flags&1 != 0 && len(l.data) > 0 {
			ofs := len(userdata)
			userdata = append(userdata, l.data...)
			ld = append(ld, userdata[ofs:])
		}
	}
	return ld
}

func (f *file) buildFrames(framesr []image.Rectangle, userdata []byte) ([]Frame, []byte) {
	frames := make([]Frame, len(f.frames))

	for i, fr := range f.frames {
		frames[i].Duration = fr.dur
		frames[i].Bounds = framesr[i]
		frames[i].Data = make([][]byte, 0, len(fr.cels))
		for _, c := range fr.cels {
			if nd := len(c.data); nd > 0 {
				ofs := len(userdata)
				userdata = append(userdata, c.data...)
				frames[i].Data = append(frames[i].Data, userdata[ofs:])
			}
		}
	}

	return frames, userdata
}

func getBounds(nframes, framew, frameh int) (atlasRect image.Rectangle, framesRects []image.Rectangle) {
	fw, fh := factorPowerOfTwo(nframes)
	if framew > frameh {
		fw, fh = fh, fw
	}

	atlasRect = image.Rect(0, 0, fw*framew, fh*frameh)

	for i := 0; i < nframes; i++ {
		x, y := i%fw, i/fw
		framesRects = append(framesRects, image.Rectangle{
			Min: image.Pt(x*framew, y*frameh),
			Max: image.Pt((x+1)*framew, (y+1)*frameh),
		})
	}

	return
}

// factorPowerOfTwo computes n=a*b, where a, b are powers of two and a >= b.
func factorPowerOfTwo(n int) (a, b int) {
	x := int(math.Ceil(math.Log2(float64(n))))
	a = 1 << (x - x/2)
	b = 1 << (x / 2)
	return
}
