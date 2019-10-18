package main

import (
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

type gameState int

const (
	waitingToDrop gameState = iota
	pawnDropped
	checkMate
)

type Scene struct {
	canvas *imdraw.IMDraw
}

func (s Scene) show(t pixel.Target) {
	s.canvas.Draw(t)
}

type Player struct {
	name string
	pawn *pixel.Sprite
}

type Block struct {
	row, col   int
	capturedBy *Player
}

func (b Block) print(s Scene) {
	padding := 2 // will be doubled for cell-cell gaps
	s.canvas.Color = colornames.Coral
	s.canvas.EndShape = imdraw.RoundEndShape
	s.canvas.Push(
		pixel.V(float64(b.col*100-padding), float64(b.row*100-padding)),         // Top-Right
		pixel.V(float64((b.col-1)*100+padding), float64((b.row-1)*100+padding)), // Bottom-Left
	)

	s.canvas.Rectangle(1)
}

func (b Block) Center() pixel.Vec {
	return pixel.V(float64(b.col*100-50), float64(b.row*100-50))
}

func (b Block) String() string {
	return strconv.Itoa(b.row) + "x" + strconv.Itoa(b.col)
}
