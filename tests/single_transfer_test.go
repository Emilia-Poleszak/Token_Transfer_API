package tests

import (
	"context"
	"testing"
	"os"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func create_test_wallets(DB *gorm.DB, t *testing.T) (models.Wallet, models.Wallet) {
	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(800)}
	toWallet := models.Wallet{Address: "0xto_address", Balance: int32(300)}

	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")
	assert.NoError(t, DB.Create(&toWallet).Error, "Creating to wallet failed")

	return fromWallet, toWallet
}

func Test_Single_Accepted_Transfer(t *testing.T) {
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
	
	fromWallet, toWallet := create_test_wallets(DB, t)
	resolver := &graph.Resolver{DB: DB}
	amount := int32(100)

	assert.Equal(t, fromWallet.Balance, int32(800), "Initial from wallet balance incorrect")
	assert.Equal(t, toWallet.Balance, int32(300), "Initial to wallet balance incorrect")

	balance, err1 := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet.Address, amount)
	assert.Equal(t, int32(700), balance, "Returned balance incorrect")
	assert.NoError(t, err1, "Transfer failed")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, updatedFromWallet.Balance, int32(700), "From wallet balance incorrect")
	assert.Equal(t, updatedToWallet.Balance, int32(400), "To wallet balance incorrect")
	
	err2 := DB.Unscoped().Where("1=1").Delete(&models.Wallet{}).Error 
	assert.NoError(t, err2)
}	

func Test_Single_Rejected_Transfer(t *testing.T) {
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

	fromWallet, toWallet := create_test_wallets(DB, t)
	resolver := &graph.Resolver{DB: DB}
	amount := int32(1000)

	assert.Equal(t, fromWallet.Balance, int32(800), "Initial from wallet balance incorrect")
	assert.Equal(t, toWallet.Balance, int32(300), "Initial to wallet balance incorrect")

	_, err1 := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet.Address, amount)
	assert.Error(t, err1, "Transfer should fail")
	assert.Contains(t, err1.Error(), "Insufficient balance")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, int32(800), updatedFromWallet.Balance, "From wallet balance should be unchanged")
	assert.Equal(t, int32(300), updatedToWallet.Balance, "To wallet balance should be unchanged")

	err2 := DB.Unscoped().Where("1=1").Delete(&models.Wallet{}).Error 
	assert.NoError(t, err2)
}