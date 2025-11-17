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

func setup_test_DB(t *testing.T) *gorm.DB {
	DB := db.ConnectDB()

	if err := DB.AutoMigrate(&models.Wallet{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	return DB
}

func create_test_wallets(DB *gorm.DB, t *testing.T) (models.Wallet, models.Wallet) {
	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(800)}
	toWallet := models.Wallet{Address: "0xto_address", Balance: int32(300)}

	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")
	assert.NoError(t, DB.Create(&toWallet).Error, "Creating to wallet failed")

	return fromWallet, toWallet
}

func Test_Single_Accepted_Transfer(t *testing.T) {
	DB := setup_test_DB(t)
	fromWallet, toWallet := create_test_wallets(DB, t)
	resolver := &graph.Resolver{DB: DB}
	amount := int32(100)

	assert.Equal(t, fromWallet.Balance, int32(800), "Initial from wallet balance incorrect")
	assert.Equal(t, toWallet.Balance, int32(300), "Initial to wallet balance incorrect")

	balance, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet.Address, amount)
	assert.Equal(t, int32(700), balance, "Returned balance incorrect")
	assert.NoError(t, err, "Transfer failed")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, updatedFromWallet.Balance, int32(700), "From wallet balance incorrect")
	assert.Equal(t, updatedToWallet.Balance, int32(400), "To wallet balance incorrect")
	
	err2 := DB.Unscoped().Where("address = ?", updatedFromWallet.Address).Delete(&models.Wallet{}).Error; 
	assert.NoError(t, err2)
	err3 := DB.Unscoped().Where("address = ?", updatedToWallet.Address).Delete(&models.Wallet{}).Error;
	assert.NoError(t, err3)
}	

func Test_Single_Rejected_Transfer(t *testing.T) {
	DB := setup_test_DB(t)
	fromWallet, toWallet := create_test_wallets(DB, t)
	resolver := &graph.Resolver{DB: DB}
	amount := int32(1000)

	assert.Equal(t, fromWallet.Balance, int32(800), "Initial from wallet balance incorrect")
	assert.Equal(t, toWallet.Balance, int32(300), "Initial to wallet balance incorrect")

	_, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet.Address, amount)
	assert.Error(t, err, "Transfer should fail")
	assert.Contains(t, err.Error(), "Insufficient balance")

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)

	assert.Equal(t, int32(800), updatedFromWallet.Balance, "From wallet balance should be unchanged")
	assert.Equal(t, int32(300), updatedToWallet.Balance, "To wallet balance should be unchanged")

	err2 := DB.Unscoped().Where("address = ?", updatedFromWallet.Address).Delete(&models.Wallet{}).Error; 
	assert.NoError(t, err2)
	err3 := DB.Unscoped().Where("address = ?", updatedToWallet.Address).Delete(&models.Wallet{}).Error;
	assert.NoError(t, err3)
}