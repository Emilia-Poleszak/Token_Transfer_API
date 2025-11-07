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

func Test_Transfer_Race_Condition(t *testing.T) {
	DB := db.ConnectDB()

	if err := DB.AutoMigrate(&models.Wallet{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	fromWallet := models.Wallet{Address: "0xfrom_address", Balance: int32(10)}
	toWallet1 := models.Wallet{Address: "0xto_address1", Balance: int32(10)}
	toWallet2 := models.Wallet{Address: "0xto_address2", Balance: int32(10)}
	toWallet3 := models.Wallet{Address: "0xto_address3", Balance: int32(10)}

	assert.NoError(t, DB.Create(&fromWallet).Error, "Creating from wallet failed")
	assert.NoError(t, DB.Create(&toWallet1).Error, "Creating to wallet 1 failed")
	assert.NoError(t, DB.Create(&toWallet2).Error, "Creating to wallet 2 failed")
	assert.NoError(t, DB.Create(&toWallet3).Error, "Creating to wallet 3 failed")
	resolver := &graph.Resolver{DB: DB}

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(3)

	var ready *sync.WaitGroup = new(sync.WaitGroup)
	ready.Add(3)

	var start *sync.WaitGroup = new(sync.WaitGroup)
	start.Add(1)

	var result string = ""

	go func() {
		defer wg.Done()

		_ = DB.First(&models.Wallet{}, "address = ?", fromWallet.Address)

		ready.Done()
		start.Wait()

		_, err := resolver.Mutation().Transfer(context.Background(), toWallet1.Address, fromWallet.Address, 1)
		
		if err == nil {
			result += " +1 accepted "
		} else {
			result += " +1 rejected "
		}
	}()

	go func() {
		defer wg.Done()

		_ = DB.First(&models.Wallet{}, "address = ?", fromWallet.Address)

		ready.Done()
		start.Wait()

		_, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet2.Address, 4)
		
		if err == nil {
			result += " -4 accepted "
		} else {
			result += " -4 rejected "
		}
	}()
	go func() {
		defer wg.Done()

		_ = DB.First(&models.Wallet{}, "address = ?", fromWallet.Address)

		ready.Done()
		start.Wait()

		_, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet3.Address, 7)
		
		if err == nil {
			result += " -7 accepted "
		} else {
			result += " -7 rejected "
		}
	}()

	ready.Wait()
	start.Done()
	wg.Wait()

	t.Logf("Transfers result: %s", result)

	var updatedFromWallet, updatedToWallet1, updatedToWallet2, updatedToWallet3 models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet1.Address).First(&updatedToWallet1)
	DB.Where("address = ?", toWallet2.Address).First(&updatedToWallet2)
	DB.Where("address = ?", toWallet3.Address).First(&updatedToWallet3)
	
	assert.GreaterOrEqual(t, updatedFromWallet.Balance, int32(0), "From wallet balance should not be negative")
	assert.LessOrEqual(t, updatedFromWallet.Balance, int32(10), "From wallet balance should not exceed initial balance")
	
	DB.Unscoped().Where("address = ?", updatedFromWallet.Address).Delete(&models.Wallet{})
	DB.Unscoped().Where("address = ?", updatedToWallet1.Address).Delete(&models.Wallet{})
	DB.Unscoped().Where("address = ?", updatedToWallet2.Address).Delete(&models.Wallet{})
	DB.Unscoped().Where("address = ?", updatedToWallet3.Address).Delete(&models.Wallet{})

	assert.Error(t, DB.Where("address = ?", updatedFromWallet.Address).First(&models.Wallet{}).Error, "Cleanup failed")
	assert.Error(t, DB.Where("address = ?", updatedToWallet1.Address).First(&models.Wallet{}).Error, "Cleanup failed")
	assert.Error(t, DB.Where("address = ?", updatedToWallet2.Address).First(&models.Wallet{}).Error, "Cleanup failed")
	assert.Error(t, DB.Where("address = ?", updatedToWallet3.Address).First(&models.Wallet{}).Error, "Cleanup failed")	
}