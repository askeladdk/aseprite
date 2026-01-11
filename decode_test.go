package aseprite

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/askeladdk/aseprite/internal/require"
)

func equalPalette(a, b color.Palette) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDecode(t *testing.T) {
	for _, tt := range []struct {
		Name     string
		Filename string
		Outfile  string
		Frames   int
		Tags     int
	}{
		{
			Name:     "paletted_index_error",
			Filename: "./testfiles/index_error.ase",
			Outfile:  "index_error.png",
			Frames:   1,
			Tags:     0,
		},
		{
			Name:     "paletted",
			Filename: "./testfiles/slime_paletted.aseprite",
			Outfile:  "slime_paletted.png",
			Frames:   10,
			Tags:     2,
		},
		{
			Name:     "grayscale",
			Filename: "./testfiles/slime_grayscale.aseprite",
			Outfile:  "slime_grayscale.png",
			Frames:   10,
			Tags:     2,
		},
		{
			Name:     "blendtest",
			Filename: "./testfiles/blendtest.aseprite",
			Outfile:  "blendtest.png",
			Frames:   1,
			Tags:     0,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			f, err := os.Open(tt.Filename)
			require.NoError(t, err)
			defer f.Close()

			img, imgformat, err := image.Decode(f)
			require.NoError(t, err)

			if imgformat != "aseprite" {
				t.Fatal(imgformat)
			}

			aspr, ok := img.(*Aseprite)
			require.True(t, ok)
			require.True(t, len(aspr.Frames) == tt.Frames, "frames", len(aspr.Frames))
			require.True(t, len(aspr.Tags) == tt.Tags, "tags", len(aspr.Tags))

			out, err := os.Create(tt.Outfile)
			require.NoError(t, err)
			require.NoError(t, png.Encode(out, img))
		})
	}
}

func TestDecodeConfig(t *testing.T) {
	for _, tt := range []struct {
		Name        string
		Filename    string
		ImageConfig image.Config
	}{
		{
			Name:     "paletted",
			Filename: "./testfiles/slime_paletted.aseprite",
			ImageConfig: image.Config{
				ColorModel: &color.Palette{
					0: color.Alpha16{0},
					1: color.NRGBA{247, 231, 198, 255},
					2: color.NRGBA{214, 142, 73, 255},
					3: color.NRGBA{166, 55, 37, 255},
					4: color.NRGBA{51, 30, 80, 255},
				},
				Width:  128,
				Height: 256,
			},
		},
		{
			Name:     "grayscale",
			Filename: "./testfiles/slime_grayscale.aseprite",
			ImageConfig: image.Config{
				ColorModel: color.Gray16Model,
				Width:      128,
				Height:     256,
			},
		},
		{
			Name:     "blendtest",
			Filename: "./testfiles/blendtest.aseprite",
			ImageConfig: image.Config{
				ColorModel: color.RGBAModel,
				Width:      640,
				Height:     360,
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			f, err := os.Open(tt.Filename)
			require.NoError(t, err)
			defer f.Close()

			conf, imgformat, err := image.DecodeConfig(f)
			require.NoError(t, err)
			require.True(t, imgformat == "aseprite", "image format", imgformat)

			require.True(t, conf.Height == tt.ImageConfig.Height, "height")
			require.True(t, conf.Width == tt.ImageConfig.Width, "width")

			if pal, ok := conf.ColorModel.(color.Palette); ok {
				imgpal := *tt.ImageConfig.ColorModel.(*color.Palette)
				require.True(t, equalPalette(pal, imgpal), "palette")
			} else {
				require.True(t, conf.ColorModel == tt.ImageConfig.ColorModel, "color model")
			}
		})
	}
}
