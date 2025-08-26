package bank

import (
	"banking-app-be/components/log"

	"github.com/jinzhu/gorm"
)

type BankModuleConfig struct {
	DB *gorm.DB
}

func NewBankModuleConfig(db *gorm.DB) *BankModuleConfig {
	return &BankModuleConfig{
		DB: db,
	}
}

func (c *BankModuleConfig) MigrateTables() {
	model := &Bank{}

	err := c.DB.AutoMigrate(model).Error
	if err != nil {
		log.NewLog().Print("Auto Migrating BankTransaction ==> %s", err)
	}

}
