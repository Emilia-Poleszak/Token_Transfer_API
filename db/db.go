package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	//"github.com/Emilia-Poleszak/Token_Transfer_API/models"
)

func ConnectDB() *gorm.DB {
	// dsn example, modify according to your setup
	dsn := "host=localhost user=myuser password=mypassword dbname=mydb port=5432 sslmode=disable TimeZone=Europe/Warsaw"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	//db.AutoMigrate(&models.Wallet{})
	return db
}