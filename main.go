package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var blocks [6][7]Block
var state gameState
var pawn, droppingPawn *pixel.Sprite
var droppingV pixel.Vec
var dropTarget *Block
var player1, player2 Player
var turnOf *Player
var pawnSheet pixel.Picture
var objects *pixel.Batch
var win *pixelgl.Window

var background *pixel.Sprite

var winTitle = "Four-In-A-Row!"
var heading, subheading string
var basicAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
var tickSound, coinSound beep.StreamSeeker

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
			turnOf.pawn.Draw(win, blockM.Moved(pixel.V(mouseX, 650)))
			subheading = turnOf.name + "'s move..."
		case pawnDropped:
			droppingV.Y -= 20
			if droppingV.Y <= dropTarget.Center().Y {
				// Reached target cell. draw permenently
				dropComplete()
				matched, from, to := checkMatching(*dropTarget)
				if matched {
					declareWin(currentScene, from, to)
				} else {
					turnOf = rotateTurn()
				}
			} else {
				droppingPawn.Draw(win, blockM.Moved(droppingV))
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
		case win.JustPressed(pixelgl.KeyQ) && (state == paused || state == intro):
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

func makeGameScene() Scene {

	stage := Scene{imdraw.New(nil), pixelgl.NewCanvas(win.Bounds())}
	stage.canvas.Color = pixel.ToRGBA(colornames.Coral).Mul(pixel.Alpha(.5))
	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			//fmt.Printf("Making Block row: %d col: %d \n", row, col)
			blocks[row][col] = Block{row + 1, col + 1, nil}
			blocks[row][col].print(stage)
		}
	}
	tryOnlineBackground()

	return stage
}

func makeIntroScene() Scene {
	discs := player1.pawn.Picture()
	stage := Scene{imdraw.New(discs), pixelgl.NewCanvas(win.Bounds())}

	stage.canvas.Color = pixel.ToRGBA(colornames.Black).Mul(pixel.Alpha(.75))
	stage.canvas.Push(
		pixel.V(620, 80), // Top-Right
		pixel.V(80, 620), // Bottom-Left
	)

	stage.canvas.Rectangle(0)

	title := text.New(pixel.V(100, 550), basicAtlas)
	fmt.Fprintln(title, "Four-In-A-Row")
	title.Draw(stage.textPad, pixel.IM.Scaled(title.Orig, 4))

	desc := text.New(pixel.V(100, 500), basicAtlas)
	fmt.Fprintln(desc, "The first one to form a horizontal,\nvertical, or diagonal line of four\nof one's own discs will win.")

	fmt.Fprintln(desc, "\n-----------------------------")
	fmt.Fprintln(desc, "<Space> : Start/Pause/Resume")
	fmt.Fprintln(desc, "\n---------(Paused)------------")
	fmt.Fprintln(desc, "<R>     : Restart")
	fmt.Fprintln(desc, "<B>     : Change Background (online)")
	fmt.Fprintln(desc, "<Q>     : Quit")
	fmt.Fprintln(desc, " ")
	fmt.Fprintln(desc, "    Player1           Player2")
	desc.Draw(stage.textPad, pixel.IM.Scaled(desc.Orig, 2))

	player1.pawn.Draw(stage.canvas, blockM.Moved(pixel.V(200, 130)))
	player2.pawn.Draw(stage.canvas, blockM.Moved(pixel.V(450, 130)))

	return stage
}

func initGame() {
	pawnSheet, err := loadPicture("assets/buttons_80.png")
	panicIfError(err)
	objects = pixel.NewBatch(&pixel.TrianglesData{}, pawnSheet)

	//sprite := pixel.NewSprite(pic, pic.Bounds())
	//button2 := pixel.NewSprite(pic, pixel.R(0, 127, 100, 227))
	buttonFrames := makeSpriteMap(pawnSheet, 80, 80)
	player1 = Player{"Player 1", colornames.Whitesmoke, pixel.NewSprite(pawnSheet, buttonFrames[0])}
	player2 = Player{"Player 2", colornames.Whitesmoke, pixel.NewSprite(pawnSheet, buttonFrames[1])}

	var format beep.Format
	format, tickSound = loadMP3Sound("assets/tick.mp3")
	format, coinSound = loadMP3Sound("assets/coin.mp3")
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
}

func restartGame(s Scene) {
	objects.Clear()
	s.canvas.Clear()
	heading = ""
	subheading = ""
	tryOnlineBackground()

	s.canvas.Color = pixel.ToRGBA(colornames.Coral).Mul(pixel.Alpha(.5))
	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			//fmt.Printf("Making Block row: %d col: %d \n", row, col)
			blocks[row][col].capturedBy = nil
			blocks[row][col].print(s)
		}
	}

	turnOf = rotateTurn()
	state = waitingToDrop
}

func getMouseXInBound(win *pixelgl.Window, min, max float64) float64 {
	if win.MousePosition().X < min {
		return min
	} else if win.MousePosition().X > max {
		return max
	}

	return win.MousePosition().X
}

func getDroppingCol(mouseX float64) int {
	return int(mouseX / 100)
}

func playMove(dropCol int, win pixel.Target) {
	state = pawnDropped
	droppingPawn = turnOf.pawn
	droppingV = blocks[5][dropCol].Center()
	droppingPawn.Draw(win, blockM.Moved(droppingV))
}

func dropComplete() {
	droppingPawn.Draw(objects, blockM.Moved(dropTarget.Center()))
	speaker.Play(tickSound)
	tickSound.Seek(0)

	dropTarget.capturedBy = turnOf
	state = waitingToDrop
}

func defaultBackground() *pixel.Sprite {
	back, err := loadPicture("assets/back_1.png")
	panicIfError(err)
	return pixel.NewSprite(back, back.Bounds())
}

func findGropTarget(col int) *Block {
	for i := 0; i <= 5; i++ {
		if blocks[i][col].capturedBy == nil {
			return &blocks[i][col]
		}
	}
	return nil
}

func rotateTurn() *Player {
	if *turnOf == player1 {
		return &player2
	}

	return &player1
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

func declareWin(s Scene, from, to Block) {
	s.canvas.Color = turnOf.color
	s.canvas.Push(from.Center(), to.Center())
	s.canvas.Line(5)
	state = checkMate
	speaker.Play(coinSound)
	coinSound.Seek(0)

	heading = turnOf.name + " Wins!"
	subheading = "Press <SPACE> to start over"
}

func tryOnlineBackground() {
	go func() {
		if onlineBackground := loadUnsplashBackground(); onlineBackground != nil {
			background = onlineBackground
		}
	}()
}

func alert(message string) {
	// println(message)
	win.SetTitle(fmt.Sprintf("!!! %s !!!", message))
	time.AfterFunc(time.Second*4, func() { win.SetTitle(winTitle) })
}

func main() {
	pixelgl.Run(run)
}

//  ------------- CHeck Matching --------------
func checkMatching(block Block) (bool, Block, Block) {
	from, to := Block{}, Block{}

	var getLastBlock = func(block Block, direction string) Block {

		directionStr := "|" + direction + "|"
		for {
			row, col := block.row, block.col

			if strings.Contains("|right|top-right|bottom-right|", directionStr) {
				col++
			}
			if strings.Contains("|left|top-left|bottom-left|", directionStr) {
				col--
			}
			if strings.Contains("|top|top-right|top-left|", directionStr) {
				row++
			}
			if strings.Contains("|bottom|bottom-left|bottom-right|", directionStr) {
				row--
			}

			if row > 6 || row < 1 || col > 7 || col < 1 || blockByRowCol(row, col).capturedBy != block.capturedBy {
				// Keep note of last 2 results by this closure
				// Also, remember to call this 2 times (from-to) for every checking
				from, to = to, block
				return block
			}

			block = blockByRowCol(row, col)
		}
	}

	switch {
	case getLastBlock(block, "right").col-getLastBlock(block, "left").col >= 3,
		getLastBlock(block, "top").row-getLastBlock(block, "bottom").row >= 3,
		getLastBlock(block, "top-right").col-getLastBlock(block, "bottom-left").col >= 3,
		getLastBlock(block, "bottom-right").col-getLastBlock(block, "top-left").col >= 3:
		return true, from, to
	}

	return false, from, to
}
