package main

import (
	"strings"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
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

var blockM = pixel.IM.Scaled(pixel.ZV, .9)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Four-In-A-Row!",
		Bounds: pixel.R(0, 0, 700, 700),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	initGame()
	background := makeBackground()
	currentScene := makeGameScene()
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
		case pawnDropped:
			droppingV.Y -= 20
			if droppingV.Y <= dropTarget.Center().Y {
				// Reached target cell. draw permenently
				dropComplete()
				checkMatching(*dropTarget)
			} else {
				droppingPawn.Draw(win, blockM.Moved(droppingV))
			}
		}

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			if state == waitingToDrop {
				dropCol := getDroppingCol(mouseX)
				dropTarget = findGropTarget(dropCol)

				if dropTarget != nil {
					playMove(dropCol, win)
				} else {
					alert("No block is left in this column!")
				}
			}
		}

		win.Update()
	}
}

func makeGameScene() Scene {

	stage := Scene{imdraw.New(nil)}

	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			//fmt.Printf("Making Block row: %d col: %d \n", row, col)
			blocks[row][col] = Block{row + 1, col + 1, nil}
			blocks[row][col].print(stage)
		}
	}

	return stage
}

func initGame() {
	pawnSheet, err := loadPicture("assets/buttons_tb.png")
	panicIfError(err)
	objects = pixel.NewBatch(&pixel.TrianglesData{}, pawnSheet)

	//sprite := pixel.NewSprite(pic, pic.Bounds())
	//button2 := pixel.NewSprite(pic, pixel.R(0, 127, 100, 227))
	buttonFrames := makeSpriteMap(pawnSheet, 100, 100)
	player1 = Player{"Ayan", pixel.NewSprite(pawnSheet, buttonFrames[0])}
	player2 = Player{"Anas", pixel.NewSprite(pawnSheet, buttonFrames[1])}
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
	dropTarget.capturedBy = turnOf
	turnOf = rotateTurn()
	state = waitingToDrop
}

func makeBackground() *pixel.Sprite {
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

func alert(message string) {
	println(message)
}

func main() {
	pixelgl.Run(run)
}

//  ------------- CHeck Matching --------------
func checkMatching(block Block) {

	switch {
	case getLastBlock(block, "right").col-getLastBlock(block, "left").col >= 3,
		getLastBlock(block, "top").row-getLastBlock(block, "bottom").row >= 3,
		getLastBlock(block, "top-right").col-getLastBlock(block, "bottom-left").col >= 3,
		getLastBlock(block, "bottom-right").col-getLastBlock(block, "top-left").col >= 3:
		alert("Yes!")
	}
}

func getLastBlock(block Block, direction string) Block {
	directionStr := "|" + direction + "|"

	for {
		row, col := block.row, block.col

		if strings.Contains("|right|top-right|bottom-right|", directionStr) {
			col++
			// fmt.Println(directionStr + " Matched |right|top-right|bottom-right|")
		}
		if strings.Contains("|left|top-left|bottom-left|", directionStr) {
			col--
			// fmt.Println(directionStr + " Matched |left|top-left|bottom-left|")
		}
		if strings.Contains("|top|top-right|top-left|", directionStr) {
			row++
			// fmt.Println(directionStr + " Matched |top|top-right|top-left|")
		}
		if strings.Contains("|bottom|bottom-left|bottom-right|", directionStr) {
			row--
			// fmt.Println(directionStr + " Matched |bottom|bottom-right|bottom-left|")
		}

		// Index in [][]blocks is -1 with row, col property
		if row > 6 || row < 1 || col > 7 || col < 1 || blocks[row-1][col-1].capturedBy != block.capturedBy {
			return block
		}

		// Index in [][]blocks is -1 with row, col property
		block = blocks[row-1][col-1]
	}
}
