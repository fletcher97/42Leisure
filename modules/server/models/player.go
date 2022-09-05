package models

import "gorm.io/gorm"

const (
	Player_loaded MsgType = iota
	Player_created
)

type Player struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;not null"`
	Stats []GameStats
}

type GameStats struct {
	gorm.Model
	PlayerID uint `gorm:"not null"`

	Name    string `gorm:"not null"`
	Wins    uint32
	Losses  uint32
	Draws   uint32
	GiveUps uint32
}

func (p1 Player) Equals(p2 Player) bool {
	return p1.ID == p2.ID
}
