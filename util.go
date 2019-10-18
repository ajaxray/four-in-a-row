package main

import (
	"image"
	"os"

	"github.com/faiface/pixel"
)

// makeSpriteMap makes slice of sprites from a spritesheet
func makeSpriteMap(spritesheet pixel.Picture, frameWidth, frameHeight float64) []pixel.Rect {
	var itemFrames []pixel.Rect
	for x := spritesheet.Bounds().Min.X; x < spritesheet.Bounds().Max.X; x += frameWidth {
		for y := spritesheet.Bounds().Min.Y; y < spritesheet.Bounds().Max.Y; y += frameHeight {
			itemFrames = append(itemFrames, pixel.R(x, y, x+frameWidth, y+frameHeight))
		}
	}

	return itemFrames
}

// loadPicture Loads picture data from path
func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
