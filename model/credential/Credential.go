package credential

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/util"
	model "banking-app-be/model/general"

	"github.com/google/uuid"
)

type Credential struct {
	model.Base
	Email    string    `json:"email" gorm:"unique;not null;type:varchar(36)"`
	Password string    `json:"password" gorm:"not null;type:varchar(255)"`
	UserID   uuid.UUID `json:"userId" gorm:"not null;type:varchar(36)"`
}

type CredentialDTO struct {
	model.Base
	Email    string    `json:"email" gorm:"unique;not null;type:varchar(36)"`
	Password string    `json:"password" gorm:"not null;type:varchar(255)"`
	UserID   uuid.UUID `json:"userId" gorm:"not null;type:varchar(36)"`
	// user user.User `json:"user"`
}

func (user *Credential) Validate() error {

	if util.IsEmpty(user.Email) || !util.ValidateEmail(user.Email) {
		return errors.NewValidationError("User Email must be specified and should be of the type abc@domain.com")
	}
	if util.IsEmpty(user.Password) || len(user.Password) < 8 {
		return errors.NewValidationError("Password should consist of 8 or more characters")
	}
	return nil
}
