package main

import (
	"fmt"
	"math/rand"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func runTest() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 700, 500),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Darkolivegreen)
	//win.SetSmooth(true)

	spritesheet, err := loadPicture("assets/buttons_tb.png")
	if err != nil {
		panic(err)
	}
	batch := pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)

	//sprite := pixel.NewSprite(pic, pic.Bounds())
	//button2 := pixel.NewSprite(pic, pixel.R(0, 127, 100, 227))
	buttonFrames := makeSpriteMap(spritesheet, 100, 100)
	button1 := pixel.NewSprite(spritesheet, buttonFrames[0])
	button2 := pixel.NewSprite(spritesheet, buttonFrames[1])

	// mat := pixel.IM
	// mat = mat.Moved(win.Bounds().Center())
	// //mat = mat.Rotated(win.Bounds().Center(), 90)
	// mat = mat.Moved(pixel.V(100, 0))
	// //mat = mat.Rotated(win.Bounds().Center(), 180)

	// button2.Draw(win, mat)

	angle := 0.0
	last := time.Now()

	var (
		trees    []*pixel.Sprite
		matrices []pixel.Matrix
		frames   = 0
		second   = time.Tick(time.Second)
	)

	last = time.Now()
	for !win.Closed() {

		dt := time.Since(last).Seconds()

		// Rotating button angle
		angle += 3 * dt

		// Make new button sprites on

		tree := pixel.NewSprite(spritesheet, buttonFrames[rand.Intn(len(buttonFrames))])
		trees = append(trees, tree)
		matrices = append(matrices, pixel.IM.Scaled(pixel.ZV, .5).Moved(win.MousePosition()))
	}

	win.Clear(colornames.Black)
	button1.Draw(win, pixel.IM.Scaled(pixel.ZV, 2).Moved(win.Bounds().Center()))

	mat := pixel.IM
	mat = mat.Rotated(pixel.ZV, angle)
	mat = mat.Moved(win.Bounds().Center())
	button2.Draw(win, mat)

	batch.Clear()
	for i, tree := range trees {
		tree.Draw(batch, matrices[i])
	}
	batch.Draw(win)

	win.Update()
	frames++
	select {
	case <-second:
		win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
		frames = 0
	default:
	}
}

// func main() {
// 	pixelgl.Run(runTest)
// }
