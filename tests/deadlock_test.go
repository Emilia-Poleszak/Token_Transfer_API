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

func Test_Deadlock_Prevention(t *testing.T) {
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
	
	err2 := DB.Unscoped().Where("1=1").Delete(&models.Wallet{}).Error 
	assert.NoError(t, err2)
}