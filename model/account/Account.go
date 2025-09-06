package account

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/util"
	model "banking-app-be/model/general"
	"banking-app-be/model/passbook"

	uuid "github.com/satori/go.uuid"
)

type Account struct {
	model.Base
	AccountNo      string                 `json:"accountNo" gorm:"unique;not null;type:varchar(20)"`
	AccountBalance float32                `json:"balance" gorm:"type:float;DEFAULT:0"`
	IsActive       *bool                  `json:"isActive" gorm:"type:tinyint(1);default:true"`
	BankID         uuid.UUID              `json:"bankId" gorm:"not null;type:varchar(36)"`
	UserID         uuid.UUID              `json:"userId" gorm:"not null;type:varchar(36)"`
	PassBook       []passbook.Transaction `json:"passbook" gorm:"foreignKey:AccountID;references:ID"`
}

type AccountDTO struct {
	model.Base
	AccountNo      string                 `json:"accountNo" gorm:"unique;not null;type:varchar(20)"`
	AccountBalance float32                `json:"balance" gorm:"type:float;DEFAULT:0"`
	IsActive       *bool                  `json:"isActive" gorm:"type:tinyint(1);default:true"`
	BankID         uuid.UUID              `json:"bankId"`
	UserID         uuid.UUID              `json:"userId"`
	User           AccountUser            `json:"user" gorm:"foreignKey:UserID"`
	PassBook       []passbook.Transaction `json:"passBook" gorm:"foreignKey:AccountID;references:ID"`
	// Bank           AccountBank            `json:"bank" gorm:"foreignKey:BankID"`
}
type AccontBankDTO struct {
	model.Base
	AccountNo      string                 `json:"accountNo" gorm:"unique;not null;type:varchar(20)"`
	AccountBalance float32                `json:"balance" gorm:"type:float;DEFAULT:0"`
	IsActive       *bool                  `json:"isActive" gorm:"type:tinyint(1);default:true"`
	BankID         uuid.UUID              `json:"bankId"`
	Bank           AccountBank            `json:"bank" gorm:"foreignKey:BankID"`
	UserID         uuid.UUID              `json:"userId"`
	PassBook       []passbook.Transaction `json:"passBook" gorm:"foreignKey:AccountID;references:ID"`
	// User           AccountUser            `json:"user" gorm:"foreignKey:UserID"`
}
type AccountUser struct {
	model.Base
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (*AccountUser) TableName() string {
	return "users"
}

type AccountBank struct {
	model.Base
	FullName string `json:"fullName"`
}

func (*AccountBank) TableName() string {
	return "banks"
}

func (*AccountDTO) TableName() string {
	return "accounts"
}

func (*AccontBankDTO) TableName() string {
	return "accounts"
}

func (a *Account) Validate() error {
	if util.IsEmpty(a.AccountNo) {
		return errors.NewValidationError("Account number must not be empty")
	}
	return nil
}
