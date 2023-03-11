// Package aseprite implements a decoder for Aseprite sprite files.
//
// Layers are flattened, blending modes are applied,
// and frames are arranged on a single texture atlas.
// Invisible and reference layers are ignored.
// Tilesets and external files are not supported.
//
// Aseprite file format spec: https://github.com/aseprite/aseprite/blob/main/docs/ase-file-specs.md
package aseprite

import (
	"image"
	"image/color"
	"io"
	"time"
)

// LoopDirection enumerates all loop animation directions.
type LoopDirection uint8

const (
	Forward LoopDirection = 0
	Reverse
	PingPong
	PingPongReverse
)

// Tag is an animation tag.
type Tag struct {
	// Name is the name of the tag. Can be duplicate.
	Name string

	// Lo is the first frame in the animation.
	Lo uint16

	// Hi is the last frame in the animation.
	Hi uint16

	// Repeat specifies how many times to repeat the animation.
	Repeat uint16

	// LoopDirection is the looping direction of the animation.
	LoopDirection LoopDirection
}

// Frame represents a single frame in the sprite.
type Frame struct {
	// Bounds is the image bounds of the frame in the sprite's atlas.
	Bounds image.Rectangle

	// Duration is the time in seconds that the frame should be displayed for
	// in a tag animation loop.
	Duration time.Duration

	// Data lists all optional user data set in the cels that make up the frame.
	// The data of invisible and reference layers is not included.
	Data [][]byte
}

// Slice represents a single slice.
type Slice struct {
	// Bounds is the bounds of the image in the texture atlas.
	Bounds image.Rectangle

	// Center is the 9-slices center relative to Bounds.
	Center image.Rectangle

	// Pivot is the pivot point relative to Bounds.
	Pivot image.Point

	// Name is the name of the slice. Can be duplicate.
	Name string

	// Data is optional user data.
	Data []byte

	// Color is the slice color.
	Color color.Color
}

// Aseprite holds the results of a parsed Aseprite image file.
type Aseprite struct {
	// Image contains all frame images in a single image.
	// Frame bounds specify where the frame images are located.
	image.Image

	// Frames lists all frames that make up the sprite.
	Frames []Frame

	// Tags lists all animation tags.
	Tags []Tag

	// Slices lists all slices.
	Slices []Slice

	// LayerData lists the user data of all visible layers.
	LayerData [][]byte
}

func (spr *Aseprite) readFrom(r io.Reader) error {
	var f file

	if _, err := f.ReadFrom(r); err != nil {
		return err
	}

	f.initPalette()

	if err := f.initLayers(); err != nil {
		return err
	}

	if err := f.initCels(); err != nil {
		return err
	}

	var framesr []image.Rectangle
	spr.Image, framesr = f.buildAtlas()
	userdata := f.buildUserData()
	spr.Frames, userdata = f.buildFrames(framesr, userdata)
	spr.LayerData = f.buildLayerData(userdata)
	spr.Tags = f.buildTags()
	spr.Slices = f.buildSlices()
	return nil
}
