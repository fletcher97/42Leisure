package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Game

type GameOption int8

const (
	GameOptionEmpty GameOption = iota
	GameOptionO
	GameOptionX
	GameOptionDraw
	GameOptionBlocked
	GameOptionStart
)

func (opt *GameOption) Scan(value interface{}) error {
	val, ok := value.(int8)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal GameOption value:", value))
	}
	*opt = GameOption(val)
	return nil
}

func (opt GameOption) Value() (driver.Value, error) {
	return int8(opt), nil
}

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

	First     GameOption `gorm:"type:smallint"`
	Second    GameOption `gorm:"type:smallint"`
	NextP     GameOption `gorm:"type:smallint"`
	PlayBoard uint8
	BigBoard  Board      `gorm:"type:smallint[]"`
	Boards    [9]Board   `gorm:"type:smallint[][]"`
	State     GameOption `gorm:"type:smallint"`
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
