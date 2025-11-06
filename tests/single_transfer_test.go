package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"
	"github.com/Emilia-Poleszak/Token_Transfer_API/db"

	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	DB := db.ConnectDB()

	if err := DB.AutoMigrate(&models.Wallet{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	return DB
}

func CreateTestWallets(DB *gorm.DB, t *testing.T) (models.Wallet, models.Wallet) {
	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(800)}
	toWallet := models.Wallet{Address: "0xto_address", Balance: int32(300)}

	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")
	assert.NoError(t, DB.Create(&toWallet).Error, "Creating to wallet failed")

	return fromWallet, toWallet
}

func TestSingleAcceptedTransfer(t *testing.T) {
	DB := setupTestDB(t)
	fromWallet, toWallet := CreateTestWallets(DB, t)
	resolver := &graph.Resolver{DB: DB}
	amount := int32(100)

	_, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet.Address, amount)
	assert.NoError(t, err, "Transfer failed")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, updatedFromWallet.Balance, int32(700), "From wallet balance incorrect")
	assert.Equal(t, updatedToWallet.Balance, int32(400), "To wallet balance incorrect")

	assert.NoError(t, DB.Where("address = ?", fromWallet.Address).Delete(&updatedFromWallet).Error, "Cleanup failed")
	assert.NoError(t, DB.Where("address = ?", toWallet.Address).Delete(&updatedToWallet).Error, "Cleanup failed")
}	