package passbook

import (
	model "banking-app-be/model/general"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	model.Base
	TimeStamp      time.Time `json:"timeStamp" gorm:"not null;type:timestamp"`
	Type           string    `json:"type" gorm:"not null;type:varchar(36)" example:"CREDIT/DEBIT"`
	Amount         float32   `json:"amount" gorm:"type:float"`
	AccountBalance float32   `json:"balance" gorm:"type:float"`
	Note           string    `json:"note" gorm:"type:varchar(100)"`
	AccountID      uuid.UUID `json:"accountId" gorm:"not null;type:varchar(36)"`
}
