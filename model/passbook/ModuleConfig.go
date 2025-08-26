package passbook

import (
	"banking-app-be/components/log"

	"github.com/jinzhu/gorm"
)

type PassbookModuleConfig struct {
	DB *gorm.DB
}

func NewPassbookModuleConfig(db *gorm.DB) *PassbookModuleConfig {
	return &PassbookModuleConfig{
		DB: db,
	}
}

func (c *PassbookModuleConfig) MigrateTables() {
	model := &Transaction{}

	// Auto migrate to create the table
	err := c.DB.AutoMigrate(model).Error
	if err != nil {
		log.NewLog().Print("Auto Migrating Transaction ==> %s", err)
	}

	// Adding foreign key constraint for AccountID referencing Account(ID)
	err = c.DB.Model(model).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		log.NewLog().Print("Foreign Key: Transaction -> Account ==> %s", err)
	}
}
