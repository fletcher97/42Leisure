package db

import (
	"42Leisure/server/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var database *gorm.DB

func Db() *gorm.DB {
	if os.Getenv("DEBUG") != "false" {
		return database.Debug()
	}
	return database
}

func InitDB() {
	dsn := "host=" + os.Getenv("PG_HOST") +
		" user=" + os.Getenv("PG_USER") +
		" password=" + os.Getenv("PG_PASS") +
		" dbname=" + os.Getenv("PG_NAME") +
		" port=" + os.Getenv("PG_PORT") +
		" sslmode=disable"
	var err error
	database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	Db().AutoMigrate(
		&models.Player{},
		&models.GameStats{},
		&models.Game_ttt{},
	)
}

func Update(item interface{}) error {
	return Db().Save(item).Error
}

func SoftDelete(item interface{}) error {
	return Db().Delete(item).Error
}

func HardDelete(item interface{}) error {
	return Db().Unscoped().Delete(item).Error
}
