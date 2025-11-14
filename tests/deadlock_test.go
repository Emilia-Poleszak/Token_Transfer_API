package tests

import (
	"context"
	"sync"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"
	"github.com/Emilia-Poleszak/Token_Transfer_API/db"
)

func Test_Deadlock_Prevention(t *testing.T) {
	DB := db.ConnectDB()	
	if err := DB.AutoMigrate(&models.Wallet{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	
	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(10)}
	toWallet := models.Wallet{Address: "0xto_address", Balance: int32(10)}
	
	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")
	assert.NoError(t, DB.Create(&toWallet).Error, "Creating to wallet failed")
	
	resolver := &graph.Resolver{DB: DB}
	
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	var ready *sync.WaitGroup = new(sync.WaitGroup)
	var start *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(2)
	ready.Add(2)
	start.Add(1)
	
	// Expecting first from->to, second to->from
	go func() {
		defer wg.Done()
		ready.Done()
		start.Wait()

		balance, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet.Address, 5)
		
		assert.Equal(t, int32(5), balance, "Returned balance incorrect")
		assert.NoError(t, err, "from->to Transfer failed")
	}()
	go func() {
		defer wg.Done()
		ready.Done()	
		start.Wait()
		
		balance, err := resolver.Mutation().Transfer(context.Background(), toWallet.Address, fromWallet.Address, 5)
		
		assert.Equal(t, int32(10), balance, "Returned balance incorrect")
		assert.NoError(t, err, "to->from Transfer failed")
	}()
	
	ready.Wait()
	start.Done()
	wg.Wait()

	var updatedFromWallet, updatedToWallet models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet.Address).First(&updatedToWallet)	
	
	assert.Equal(t, int32(10), updatedFromWallet.Balance, "From wallet balance incorrect")
	assert.Equal(t, int32(10), updatedToWallet.Balance, "To wallet balance incorrect")
	
	err2 := DB.Unscoped().Where("address = ?", updatedFromWallet.Address).Delete(&models.Wallet{}).Error; 
	assert.NoError(t, err2)
	err3 := DB.Unscoped().Where("address = ?", updatedToWallet.Address).Delete(&models.Wallet{}).Error;
	assert.NoError(t, err3)
}