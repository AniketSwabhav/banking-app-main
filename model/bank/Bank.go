package bank

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/util"
	"banking-app-be/model/account"
	banktransaction "banking-app-be/model/bankTransaction"
	model "banking-app-be/model/general"
)

type Bank struct {
	model.Base
	FullName         string                            `json:"fullName" example:"State Bank of India" gorm:"type:varchar(100)"`
	Abbreviation     string                            `json:"abbreviation" example:"SBI" gorm:"type:varchar(36)"`
	IsActive         bool                              `json:"isActive" gorm:"type:boolean;default:true"`
	Accounts         []account.Account                 `json:"accounts" gorm:"foreignKey:BankID;references:ID"`
	BankTransactions []banktransaction.BankTransaction `json:"bankTransactions" gorm:"foreignKey:SenderBankID;references:ID"`
	// BankTransactions []banktransaction.BankTransaction `json:"bankTransactions" gorm":"foreignKey:SenderBankID;references:ID"`
}

func (bank *Bank) Validate() error {

	if util.IsEmpty(bank.FullName) || util.ValidateString(bank.FullName) {
		return errors.NewValidationError("Banks Full Name must be specified and must have characters only")
	}

	if util.IsEmpty(bank.Abbreviation) || util.ValidateString(bank.Abbreviation) {
		return errors.NewValidationError("Banks Abbreviation must be specified and must have characters only")
	}

	return nil
}
