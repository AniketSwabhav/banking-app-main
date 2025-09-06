package user

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/util"
	"banking-app-be/model/account"
	"banking-app-be/model/credential"
	model "banking-app-be/model/general"
)

type User struct {
	model.Base
	FirstName    string                 `json:"firstName" example:"Ravi" gorm:"type:varchar(500)"`
	LastName     string                 `json:"lastName" example:"Sharma" gorm:"type:varchar(50)"`
	PhoneNo      string                 `sql:"index" json:"phoneNo" example:"9700795509" gorm:"type:varchar(15)"`
	IsAdmin      *bool                  `json:"isAdmin" gorm:"type:tinyint(1);default:false"`
	IsActive     *bool                  `json:"isActive" gorm:"type:tinyint(1);default:true"`
	TotalBalance float32                `json:"totalBalance" gorm:"type:float;DEFAULT:0"`
	Credentials  *credential.Credential `json:"credential"`
}

type UserDTO struct {
	model.Base
	FirstName    string                    `json:"firstName" example:"Ravi" gorm:"type:varchar(50)"`
	LastName     string                    `json:"lastName" example:"Sharma" gorm:"type:varchar(50)"`
	PhoneNo      string                    `sql:"index" json:"phoneNo" example:"9700795509" gorm:"type:varchar(15)"`
	IsAdmin      *bool                     `json:"isAdmin" gorm:"type:tinyint(1);default:false"`
	IsActive     *bool                     `json:"isActive" gorm:"type:tinyint(1);default:true"`
	TotalBalance float32                   `json:"totalBalance" gorm:"type:float;DEFAULT:0"`
	Credentials  *credential.CredentialDTO `json:"credential" gorm:"foreignKey:UserId;"`
	Accounts     []account.AccontBankDTO   `json:"accounts" gorm:"foreignKey:UserId;"`
	// Accounts     []account.AccountDTO   `json:"accounts" gorm:"foreignKey:UserId;references:ID"`
}

func (*UserDTO) TableName() string {
	return "users"
}

func (user *User) Validate() error {

	if util.IsEmpty(user.FirstName) || !util.ValidateString(user.FirstName) {
		return errors.NewValidationError("User FirstName must be specified and must have characters only")
	}
	if util.IsEmpty(user.LastName) || !util.ValidateString(user.LastName) {
		return errors.NewValidationError("User LastName must be specified and must have characters only")
	}
	if util.IsEmpty(user.PhoneNo) || !util.ValidateContact(user.PhoneNo) {
		return errors.NewValidationError("User Contact must be specified and have 10 digits")
	}
	return nil
}
