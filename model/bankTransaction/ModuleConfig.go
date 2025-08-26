package banktransaction

import (
	"banking-app-be/components/log"

	"github.com/jinzhu/gorm"
)

type BankTransactionModuleConfig struct {
	DB *gorm.DB
}

func NewBankTransactionModuleConfig(db *gorm.DB) *BankTransactionModuleConfig {
	return &BankTransactionModuleConfig{
		DB: db,
	}
}

func (u *BankTransactionModuleConfig) MigrateTables() {

	model := &BankTransaction{}

	err := u.DB.AutoMigrate(model).Error
	if err != nil {
		log.NewLog().Print("Auto Migrating bankTransactions ==> %s", err)
	}

	// Foreign key constraint: SenderBankID -> Bank(ID)
	err = u.DB.Model(model).AddForeignKey("sender_bank_id", "banks(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		log.NewLog().Print("Foreign Key: BankTransaction -> SenderBank ==> %s", err)
	}

	// Foreign key constraint: ReceiverBankID -> Bank(ID)
	err = u.DB.Model(model).AddForeignKey("receiver_bank_id", "banks(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		log.NewLog().Print("Foreign Key: BankTransaction -> ReceiverBank ==> %s", err)
	}
}
