package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep/speaker"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

type gameState int

const (
	intro gameState = iota
	waitingToDrop
	pawnDropped
	paused
	checkMate
)

// Scene is a collection off-screen targets that presents a game scene
type Scene struct {
	canvas  *imdraw.IMDraw
	textPad *pixelgl.Canvas
}

func (s Scene) show(t pixel.Target) {
	s.canvas.Draw(t)
	s.textPad.Draw(t, pixel.IM.Moved(win.Bounds().Center()))
}

// Player represent each participents of game
type Player struct {
	name  string
	color color.RGBA
	disc  *pixel.Sprite
}

// Block is a cell of game board that can hold a disc
type Block struct {
	row, col   int
	capturedBy *Player
}

func (b Block) print(s Scene) {
	padding := 2 // will be doubled for cell-cell gaps
	s.canvas.Push(
		pixel.V(float64(b.col*100-padding), float64(b.row*100-padding)),         // Top-Right
		pixel.V(float64((b.col-1)*100+padding), float64((b.row-1)*100+padding)), // Bottom-Left
	)

	s.canvas.Rectangle(1)
}

// Center finds the center point (Vector) of a Block
func (b Block) Center() pixel.Vec {
	return pixel.V(float64(b.col*100-50), float64(b.row*100-50))
}

func (b Block) String() string {
	return strconv.Itoa(b.row) + "x" + strconv.Itoa(b.col)
}

// blockByRowCol finds the block in blocks
// Index in [][]blocks is -1 with row, col property
func blockByRowCol(row, col int) Block {
	return blocks[row-1][col-1]
}

// ============== Prepare different game scene ===========
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
	discs := player1.disc.Picture()
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

	player1.disc.Draw(stage.canvas, blockM.Moved(pixel.V(200, 130)))
	player2.disc.Draw(stage.canvas, blockM.Moved(pixel.V(450, 130)))

	return stage
}

// ------------------------------------------------------------

// ============== Display game states and events =========================
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

// Drop the disc in appropriate column
func playMove(dropCol int, win pixel.Target) {
	state = pawnDropped
	droppingDisc = turnOf.disc
	droppingV = blocks[5][dropCol].Center()
	droppingDisc.Draw(win, blockM.Moved(droppingV))
}

// Disc reached to it's target block
func dropComplete() {
	droppingDisc.Draw(objects, blockM.Moved(dropTarget.Center()))
	speaker.Play(tickSound)
	tickSound.Seek(0)

	dropTarget.capturedBy = turnOf
	state = waitingToDrop
}

func rotateTurn() *Player {
	if *turnOf == player1 {
		return &player2
	}

	return &player1
}

func declareWin(s Scene, from, to Block) {
	s.canvas.Color = turnOf.color
	s.canvas.Push(from.Center(), to.Center())
	s.canvas.Line(5)
	state = checkMate
	speaker.Play(coinSound)
	coinSound.Seek(0)

	heading = turnOf.name + " Wins!"
	subheading = "Press <SPACE> to start, <Q> to quit."
}

// -------------------------------------------------

//  ------------- CHeck Line Matching --------------
func getLastBlock(block Block, direction string) Block {

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
			return block
		}

		block = blockByRowCol(row, col)
	}
}

func checkMatching(block Block) (bool, Block, Block) {
	from, to := Block{}, Block{}

	from = getLastBlock(block, "right")
	to = getLastBlock(block, "left")
	if from.col-to.col >= 3 {
		return true, from, to
	}

	from = getLastBlock(block, "top")
	to = getLastBlock(block, "bottom")
	if from.row-to.row >= 3 {
		return true, from, to
	}

	from = getLastBlock(block, "top-right")
	to = getLastBlock(block, "bottom-left")
	if from.col-to.col >= 3 {
		return true, from, to
	}

	from = getLastBlock(block, "bottom-right")
	to = getLastBlock(block, "top-left")
	if from.col-to.col >= 3 {
		return true, from, to
	}

	return false, from, to
}

// ============== Utilities and calcuations =========================
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

func alert(message string) {
	// println(message)
	win.SetTitle(fmt.Sprintf("!!! %s !!!", message))
	time.AfterFunc(time.Second*4, func() { win.SetTitle(winTitle) })
}

// ------------------------------------------------------------
