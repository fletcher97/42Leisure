package models

import (
	"gorm.io/gorm"
)

// Game

type GameOption int

const (
	GameOptionEmpty GameOption = iota
	GameOptionO
	GameOptionX
	GameOptionDraw
	GameOptionBlocked
	GameOptionStart
)

const (
	First  = GameOptionO
	Second = GameOptionX
)

type Board [9]GameOption

type Game_ttt struct {
	gorm.Model

	P1ID uint
	P1   Player
	P2ID uint
	P2   Player

	First     GameOption
	Second    GameOption
	NextP     GameOption
	PlayBoard uint8
	BigBoard  Board    `gorm:"type:smallint[]"`
	Boards    [9]Board `gorm:"type:smallint[][]"`
	State     GameOption
}

// Com

type MsgType int8

const (
	Ttt_join MsgType = iota
	Ttt_create
	Ttt_list
	Ttt_quit
	Ttt_giveUp
	Ttt_play
	Ttt_failed
	Ttt_badMove
	Ttt_ok
)

func (t MsgType) Bytes() []byte {
	return []byte{byte(t)}
}
