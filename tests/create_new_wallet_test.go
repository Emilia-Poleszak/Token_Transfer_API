package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"
	"github.com/Emilia-Poleszak/Token_Transfer_API/db"
)

func Test_Transfer_To_New_Wallet(t *testing.T) {
	DB := db.ConnectDB()
	if err := DB.AutoMigrate(&models.Wallet{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	resolver := &graph.Resolver{DB: DB}
	
	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(800)}
	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")

	to_address := "0xto_address"
	amount := int32(100)

	assert.Equal(t, fromWallet.Balance, int32(800), "Initial from wallet balance incorrect")

	var toWallet models.Wallet
	balance, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, to_address, amount)
	err2 := DB.Where("address = ?", to_address).First(&toWallet).Error
	
	assert.Equal(t, int32(700), balance, "Returned balance incorrect")
	assert.NoError(t, err2, "New to wallet was not created")
	assert.NoError(t, err, "Transfer failed")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, updatedFromWallet.Balance, int32(700), "From wallet balance incorrect")
	assert.Equal(t, updatedToWallet.Balance, int32(100), "To wallet balance incorrect")
	
	err3 := DB.Unscoped().Where("address = ?", updatedFromWallet.Address).Delete(&models.Wallet{}).Error; 
	assert.NoError(t, err3)
	err4 := DB.Unscoped().Where("address = ?", updatedToWallet.Address).Delete(&models.Wallet{}).Error;
	assert.NoError(t, err4)
}	