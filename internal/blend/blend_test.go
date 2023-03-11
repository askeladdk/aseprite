package blend

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strings"
	"testing"

	"github.com/askeladdk/aseprite/internal/require"
)

func jpgDecode(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return jpeg.Decode(f)
}

func jpgEncode(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
}

func TestBlendModes(t *testing.T) {
	dst, err := jpgDecode("../../testfiles/dst.jpg")
	require.NoError(t, err)

	src, err := jpgDecode("../../testfiles/src.jpg")
	require.NoError(t, err)

	for i, name := range []string{
		"Normal",
		"Multiply",
		"Screen",
		"Overlay",
		"Darken",
		"Lighten",
		"ColorDodge",
		"ColorBurn",
		"HardLight",
		"SoftLight",
		"Difference",
		"Exclusion",
		"Hue",
		"Saturation",
		"Color",
		"Luminosity",
		"Addition",
		"Subtract",
		"Divide",
	} {
		t.Run(name, func(t *testing.T) {
			img := image.NewRGBA(src.Bounds())
			Blend(img, img.Bounds(), src, image.Point{}, dst, image.Point{}, Modes[i])

			require.NoError(t, jpgEncode(fmt.Sprintf("out_%s.jpg", strings.ToLower(name)), img))
		})
	}
}
