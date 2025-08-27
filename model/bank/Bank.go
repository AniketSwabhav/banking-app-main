package bank

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/util"
	"banking-app-be/model/account"
	banktransaction "banking-app-be/model/bankTransaction"
	model "banking-app-be/model/general"
	"strings"
)

type Bank struct {
	model.Base
	FullName         string                            `json:"fullName" example:"State Bank of India" gorm:"type:varchar(100);not null"`
	Abbreviation     string                            `json:"abbreviation" example:"SBI" gorm:"type:varchar(36);not null"`
	IsActive         bool                              `json:"isActive" gorm:"type:boolean;default:true"`
	Accounts         []account.Account                 `json:"-" gorm:"foreignKey:BankID;references:ID"`
	BankTransactions []banktransaction.BankTransaction `json:"-" gorm:"foreignKey:SenderBankID;references:ID"`
}

type BankDTO struct {
	model.Base
	FullName         string                            `json:"fullName"`
	Abbreviation     string                            `json:"abbreviation"`
	IsActive         bool                              `json:"isActive"`
	Accounts         []account.Account                 `json:"accounts,omitempty"`
	BankTransactions []banktransaction.BankTransaction `json:"bankTransactions,omitempty"`
}

func (*BankDTO) TableName() string {
	return "banks"
}

func (bank *Bank) Validate() error {
	if util.IsEmpty(bank.FullName) || !util.ValidateString(bank.FullName) {
		return errors.NewValidationError("Bank's Full Name must be specified and contain characters only")
	}
	if util.IsEmpty(bank.Abbreviation) {
		bank.Abbreviation = GetAbbreviation(bank.FullName)
	}
	return nil
}

func GetAbbreviation(input string) string {
	words := strings.Fields(input)
	var firstLetters []string
	for _, word := range words {
		if len(word) > 0 {
			firstLetters = append(firstLetters, string(word[0]))
		}
	}
	return strings.ToUpper(strings.Join(firstLetters, ""))
}
