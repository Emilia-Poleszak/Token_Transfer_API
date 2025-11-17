package tests

import (
	"context"
	"testing"
	"os"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"
)

func Test_Transfer_To_New_Wallet(t *testing.T) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := "tests"
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
    	"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Warsaw",
    	dbHost, dbUser, dbPass, dbName, dbPort,
	)
	
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err, "Failed to connect to database")

	automigrate_err := DB.AutoMigrate(&models.Wallet{})
	assert.NoError(t, automigrate_err, "AutoMigrate failed")

	resolver := &graph.Resolver{DB: DB}
	
	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(800)}
	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")

	to_address := "0xto_address"
	amount := int32(100)

	assert.Equal(t, fromWallet.Balance, int32(800), "Initial from wallet balance incorrect")

	var toWallet models.Wallet
	balance, err1 := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, to_address, amount)
	assert.NoError(t, err1, "Transfer failed")
	assert.Equal(t, int32(700), balance, "Returned balance incorrect")

	err2 := DB.Where("address = ?", to_address).First(&toWallet).Error	
	assert.NoError(t, err2, "New to wallet was not created")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, updatedFromWallet.Balance, int32(700), "From wallet balance incorrect")
	assert.Equal(t, updatedToWallet.Balance, int32(100), "To wallet balance incorrect")
	
	err3 := DB.Unscoped().Where("1=1").Delete(&models.Wallet{}).Error 
	assert.NoError(t, err3)
}	