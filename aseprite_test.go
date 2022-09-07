package aseprite

import (
	"image/png"
	"os"
	"testing"
)

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnimation(t *testing.T) {
	f, err := os.Open("./testfiles/slime_paletted.aseprite")
	assertNoError(t, err)
	defer f.Close()

	var spr Aseprite
	_, err = spr.ReadFrom(f)
	assertNoError(t, err)

	out, _ := os.Create("slime_paletted.png")
	assertNoError(t, png.Encode(out, spr.Atlas))
}

func TestGrayscale(t *testing.T) {
	f, err := os.Open("./testfiles/slime_grayscale.aseprite")
	assertNoError(t, err)
	defer f.Close()

	var spr Aseprite
	_, err = spr.ReadFrom(f)
	assertNoError(t, err)

	out, _ := os.Create("slime_grayscale.png")
	assertNoError(t, png.Encode(out, spr.Atlas))
}

func TestBlend(t *testing.T) {
	f, err := os.Open("./testfiles/blendtest.aseprite")
	assertNoError(t, err)
	defer f.Close()

	var spr Aseprite
	_, err = spr.ReadFrom(f)
	assertNoError(t, err)

	out, _ := os.Create("blendtest.png")
	assertNoError(t, png.Encode(out, spr.Atlas))
}
