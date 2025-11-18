package tests

import (
	"context"
	"sync"
	"testing"	
	"os"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"
)

func Test_Transfer_Race_Condition(t *testing.T) {
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

		ready.Done()
		start.Wait()

		_, err := resolver.Mutation().Transfer(context.Background(), toWallet1.Address, fromWallet.Address, 1)
		
		if err == nil {
			result += " +1 accepted,"
		} else {
			assert.Contains(t, err.Error(), "Insufficient balance")
			result += " +1 rejected,"
		}
	}()

	go func() {
		defer wg.Done()

		ready.Done()
		start.Wait()

		_, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet2.Address, 4)
		
		if err == nil {
			result += " -4 accepted,"
		} else {
			assert.Contains(t, err.Error(), "Insufficient balance")
			result += " -4 rejected,"
		}
	}()
	go func() {
		defer wg.Done()

		ready.Done()
		start.Wait()

		_, err := resolver.Mutation().Transfer(context.Background(), fromWallet.Address, toWallet3.Address, 7)
		
		if err == nil {
			result += " -7 accepted,"
		} else {
			assert.Contains(t, err.Error(), "Insufficient balance")
			result += " -7 rejected,"
		}
	}()

	ready.Wait()
	start.Done()
	wg.Wait()

	t.Logf("\nTransfers result:%s", result)

	var updatedFromWallet, updatedToWallet1, updatedToWallet2, updatedToWallet3 models.Wallet
	DB.Where("address = ?", fromWallet.Address).First(&updatedFromWallet)
	DB.Where("address = ?", toWallet1.Address).First(&updatedToWallet1)
	DB.Where("address = ?", toWallet2.Address).First(&updatedToWallet2)
	DB.Where("address = ?", toWallet3.Address).First(&updatedToWallet3)
	
	assert.GreaterOrEqual(t, updatedFromWallet.Balance, int32(0), "From wallet balance should not be negative")
	assert.LessOrEqual(t, updatedFromWallet.Balance, int32(10), "From wallet balance should not exceed initial balance")

	assert.Equal(t, updatedToWallet1.Balance, int32(9), "To wallet 1 balance incorrect")
	assert.True(t, updatedToWallet2.Balance == int32(10) || updatedToWallet2.Balance == int32(14), "To wallet 2 balance incorrect")
	assert.True(t, updatedToWallet3.Balance == int32(10) || updatedToWallet3.Balance == int32(17), "To wallet 3 balance incorrect")

	err1 := DB.Unscoped().Where("1=1").Delete(&models.Wallet{}).Error 
	assert.NoError(t, err1)
}