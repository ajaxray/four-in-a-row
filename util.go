package main

import (
	"image"
	"net/http"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
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

// loadPicture Loads picture data from path
func loadPictureURL(url string) (pixel.Picture, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

// Load the sound in buffer and returns the SoundSeeker and format
// How to use:
// format, coinSound = loadMP3Sound("assets/coin.mp3")
// speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
// speaker.Play(tickSound)
// tickSound.Seek(0)
func loadMP3Sound(filePath string) (beep.Format, beep.StreamSeeker) {
	f, err := os.Open(filePath)
	panicIfError(err)

	streamer, format, err := mp3.Decode(f)
	panicIfError(err)

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	return format, buffer.Streamer(0, buffer.Len())
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
