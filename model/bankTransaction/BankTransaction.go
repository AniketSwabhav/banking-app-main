package banktransaction

import (
	model "banking-app-be/model/general"

	uuid "github.com/satori/go.uuid"
)

type BankTransaction struct {
	model.Base
	SenderBankID   uuid.UUID `json:"senderBankId" gorm:"not null;type:varchar(36)"`
	ReceiverBankID uuid.UUID `json:"receiverBankId" gorm:"not null;type:varchar(36)"`
	Amount         float32   `json:"amount" gorm:"type:float"`
}

type BankTransactionDTO struct {
	model.Base
	SenderBankID   uuid.UUID        `json:"senderBankId" gorm:"not null;type:varchar(36)"`
	SenderBank     SenderBankName   `json:"senderBankName" gorm:"foreignKey: SenderBankID"`
	ReceiverBankID uuid.UUID        `json:"receiverBankId" gorm:"not null;type:varchar(36)"`
	ReceiverBank   ReceiverBankName `json:"receiverBankName" gorm:"foreignKey: ReceiverBankID"`
	Amount         float32          `json:"amount" gorm:"type:float"`
}

type SenderBankName struct {
	model.Base
	FullName string `json:"fullName"`
}

func (*SenderBankName) TableName() string {
	return "banks"
}

type ReceiverBankName struct {
	model.Base
	FullName string `json:"fullName"`
}

func (*ReceiverBankName) TableName() string {
	return "banks"
}

func (*BankTransactionDTO) TableName() string {
	return "bank_transactions"
}
