package account

import (
	"banking-app-be/components/log"

	"github.com/jinzhu/gorm"
)

type AccountModuleConfig struct {
	DB *gorm.DB
}

func NewAccountModuleConfig(db *gorm.DB) *AccountModuleConfig {
	return &AccountModuleConfig{
		DB: db,
	}
}

func (c *AccountModuleConfig) MigrateTables() {

	model := &Account{}

	err := c.DB.AutoMigrate(model).Error
	if err != nil {
		log.NewLog().Print("Auto Migrating Credential ==> %s", err)
	}

	// Foreign key: accounts.user_id → users.id
	err = c.DB.Model(model).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		log.NewLog().Print("Foreign Key: Account -> User ==> %s", err)
	}

	// Foreign key: accounts.bank_id → banks.id
	err = c.DB.Model(model).AddForeignKey("bank_id", "banks(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		log.NewLog().Print("Foreign Key: Account -> Bank ==> %s", err)
	}

}
