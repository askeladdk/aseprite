package aseprite

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecode(t *testing.T) {
	for _, tt := range []struct {
		Name     string
		Filename string
		Outfile  string
	}{
		{
			Name:     "paletted",
			Filename: "./testfiles/slime_paletted.aseprite",
			Outfile:  "slime_paletted.png",
		},
		{
			Name:     "grayscale",
			Filename: "./testfiles/slime_grayscale.aseprite",
			Outfile:  "slime_grayscale.png",
		},
		{
			Name:     "blendtest",
			Filename: "./testfiles/blendtest.aseprite",
			Outfile:  "blendtest.png",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			f, err := os.Open(tt.Filename)
			assertNoError(t, err)
			defer f.Close()

			img, imgformat, err := image.Decode(f)
			assertNoError(t, err)

			if imgformat != "aseprite" {
				t.Fatal(imgformat)
			}

			if _, ok := img.(*Aseprite); !ok {
				t.Fatal()
			}

			out, err := os.Create(tt.Outfile)
			assertNoError(t, err)
			assertNoError(t, png.Encode(out, img))
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
				ColorModel: color.RGBAModel,
				Width:      128,
				Height:     256,
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
			assertNoError(t, err)
			defer f.Close()

			conf, imgformat, err := image.DecodeConfig(f)
			assertNoError(t, err)

			if imgformat != "aseprite" {
				t.Fatal(imgformat)
			}

			if conf != tt.ImageConfig {
				t.Fatal(conf)
			}
		})
	}
}
