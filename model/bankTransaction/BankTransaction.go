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
