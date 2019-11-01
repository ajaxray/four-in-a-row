package main

import (
	"fmt"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"github.com/subosito/gotenv"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var blocks [6][7]Block
var state gameState
var droppingV pixel.Vec
var dropTarget *Block
var player1, player2 Player
var turnOf *Player
var objects *pixel.Batch
var win *pixelgl.Window

var background *pixel.Sprite
var disc, droppingDisc *pixel.Sprite

var winTitle = "Four-In-A-Row!"
var heading, subheading string
var basicAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
var tickSound, coinSound beep.StreamSeeker
var useUnsplash bool

//var blockM = pixel.IM.Scaled(pixel.ZV, .8)
var blockM = pixel.IM

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  winTitle,
		Bounds: pixel.R(0, 0, 700, 700),
		VSync:  true,
	}
	win, _ = pixelgl.NewWindow(cfg)

	initGame()
	background = defaultBackground()
	pauseModal := makeIntroScene()
	currentScene := pauseModal
	turnOf = &player1

	for !win.Closed() {
		win.Clear(colornames.Darkslategray)
		background.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		mouseX := getMouseXInBound(win, 50, 650)

		// drawBackground(win)
		currentScene.show(win)
		objects.Draw(win)

		switch state {
		case waitingToDrop:
			turnOf.disc.Draw(win, blockM.Moved(pixel.V(mouseX, 650)))
			subheading = turnOf.name + "'s move..."
		case pawnDropped:
			droppingV.Y -= 20
			if droppingV.Y <= dropTarget.Center().Y {
				// Reached target cell. draw permanently on game scene
				dropComplete()
				matched, from, to := checkMatching(*dropTarget)
				if matched {
					declareWin(currentScene, from, to)
				} else {
					turnOf = rotateTurn()
				}
			} else {
				droppingDisc.Draw(win, blockM.Moved(droppingV))
			}
		}

		switch {
		case win.JustPressed(pixelgl.MouseButtonLeft) && state == waitingToDrop:
			dropCol := getDroppingCol(mouseX)
			dropTarget = findGropTarget(dropCol)

			if dropTarget != nil {
				playMove(dropCol, win)
			} else {
				alert("No block is left in this column!")
			}
		case win.JustPressed(pixelgl.KeySpace) && state == intro:
			currentScene = makeGameScene()
			state = waitingToDrop
		case win.JustPressed(pixelgl.KeySpace) && state == waitingToDrop:
			state = paused
		case win.JustPressed(pixelgl.KeyQ) && (state == paused || state == intro || state == checkMate):
			win.SetClosed(true)
		case win.JustPressed(pixelgl.KeyB) && state == paused:
			tryOnlineBackground()
			state = waitingToDrop
		case win.JustPressed(pixelgl.KeyR) && state == paused:
			restartGame(currentScene)
		case win.JustPressed(pixelgl.KeySpace) && state == paused:
			state = waitingToDrop
		case win.JustPressed(pixelgl.KeySpace) && state == checkMate:
			restartGame(currentScene)
		}

		if state == paused {
			pauseModal.show(win)
		} else {
			printTitles()
		}

		win.Update()
	}
}

func initGame() {
	discSheet, err := loadPicture("assets/buttons_80.png")
	panicIfError(err)
	objects = pixel.NewBatch(&pixel.TrianglesData{}, discSheet)

	//sprite := pixel.NewSprite(pic, pic.Bounds())
	//button2 := pixel.NewSprite(pic, pixel.R(0, 127, 100, 227))
	buttonFrames := makeSpriteMap(discSheet, 80, 80)
	player1 = Player{"Player 1", colornames.Whitesmoke, pixel.NewSprite(discSheet, buttonFrames[0])}
	player2 = Player{"Player 2", colornames.Whitesmoke, pixel.NewSprite(discSheet, buttonFrames[1])}

	var format beep.Format
	format, tickSound = loadMP3Sound("assets/tick.mp3")
	format, coinSound = loadMP3Sound("assets/coin.mp3")
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
}

func printTitles() {

	if heading != "" {
		message := text.New(pixel.V(100, 650), basicAtlas)
		fmt.Fprintln(message, heading)
		message.Draw(win, pixel.IM.Scaled(message.Orig, 4))
	}

	if subheading != "" {
		message := text.New(pixel.V(100, 620), basicAtlas)
		fmt.Fprintln(message, subheading)
		message.Draw(win, pixel.IM.Scaled(message.Orig, 2))
	}
}

func init() {
	gotenv.Load()
}

func main() {
	go func() {
		useUnsplash = prepareUnsplash(
			getEnvStr("UNSPLASH_API_URL", "https://api.unsplash.com"),
			getEnvStr("UNSPLASH_ACCESS_KEY", ""),
			getEnvInt("UNSPLASH_COLLECTION_ID", 0),
		)
	}()

	pixelgl.Run(run)
}

func tryOnlineBackground() {
	if useUnsplash {
		go func() {
			if onlineBackground := loadUnsplashBackground(); onlineBackground != nil {
				background = onlineBackground
			}
		}()
	} else {
		alert("Sorry! Unsplash images are not available at this moment.")
	}
}
