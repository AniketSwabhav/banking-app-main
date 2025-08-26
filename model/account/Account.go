package account

import (
	model "banking-app-be/model/general"
	"banking-app-be/model/passbook"

	uuid "github.com/satori/go.uuid"
)

type Account struct {
	model.Base
	AccountNo      string                 `json:"accountNo" gorm:"unique;not null;type:varchar(20)"`
	AccountBalance float32                `json:"balance" gorm:"type:float;DEFAULT:0"`
	IsActive       bool                   `json:"isActive"`
	BankID         uuid.UUID              `json:"bankId" gorm:"not null;type:varchar(36)"`
	UserID         uuid.UUID              `json:"userId" gorm:"not null;type:varchar(36)"`
	PassBook       []passbook.Transaction `json:"passBook" gorm:"foreignKey:AccountID;references:ID"`
}
