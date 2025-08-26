package service

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/model/bank"
	"banking-app-be/module/repository"

	"github.com/jinzhu/gorm"
)

const cost = 10

type BankService struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewBankService(DB *gorm.DB, repo repository.Repository) *BankService {
	return &BankService{
		db:         DB,
		repository: repo,
	}
}

func (service *BankService) CreateBank(newBank *bank.Bank) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	if err := newBank.Validate(); err != nil {
		log.GetLogger().Error(err.Error())
		uow.RollBack()
		return err
	}

	err := uow.DB.Create(newBank).Error
	if err != nil {
		return errors.NewDatabaseError("Failed to create user")
	}

	uow.Commit()
	return nil
}
