package models

import "gorm.io/gorm"

const (
	Player_loaded MsgType = iota
	Player_created
)

type Player struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex"`
	Stats []GameStats
}

type GameStats struct {
	gorm.Model
	PlayerID uint

	Name    string
	Wins    uint32
	Losses  uint32
	Draws   uint32
	GiveUps uint32
}

func (p1 Player) Equals(p2 Player) bool {
	return p1.ID == p2.ID
}
