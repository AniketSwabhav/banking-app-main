package service

import (
	"banking-app-be/components/errors"
	"banking-app-be/model/account"
	"banking-app-be/model/bank"
	"banking-app-be/model/passbook"
	"banking-app-be/model/user"
	"banking-app-be/module/repository"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type PassbookService struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewPassbookService(DB *gorm.DB, repo repository.Repository) *PassbookService {
	return &PassbookService{
		db:         DB,
		repository: repo,
	}
}

func (service *PassbookService) GetPassbookByAccountNo(passbook *[]passbook.Transaction, userId uuid.UUID, accountNo *string, totalCount *int, limit, offset int) error {
	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	accountOwner := user.User{}
	if err := service.repository.GetRecordByID(uow, userId, &accountOwner); err != nil {
		return errors.NewDatabaseError("user not found")
	}
	if accountOwner.IsActive != nil && !*accountOwner.IsActive {
		return errors.NewInActiveUserError("can not get the passbook records for InActive user")
	}

	userAccount := account.Account{}
	if err := service.repository.GetRecord(uow, &userAccount, repository.Filter("account_no = ?", accountNo)); err != nil {
		return errors.NewValidationError("Record not found for given Account number")
	}
	if userAccount.IsActive != nil && !*userAccount.IsActive {
		return errors.NewInActiveUserError("can not get the passbook records for InActive Account")
	}

	accountOwnerBank := bank.Bank{}
	if err := service.repository.GetRecordByID(uow, userAccount.BankID, &accountOwnerBank); err != nil {
		return errors.NewValidationError("Bank not found")
	}
	if accountOwnerBank.IsActive != nil && !*accountOwnerBank.IsActive {
		return errors.NewInActiveUserError("can not get the passbook records for InActive Bank")
	}

	queryProcessor := []repository.QueryProcessor{
		repository.Filter("account_id = ?", userAccount.ID),
		repository.Paginate(limit, offset, totalCount),
	}
	err := service.repository.GetAll(uow, passbook, queryProcessor...)
	if err != nil {
		return err
	}

	err = service.repository.GetCount(uow, passbook, totalCount, repository.Filter("account_id = ?", userAccount.ID))
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (service *PassbookService) GetPassbookByAccountId(passbook *[]passbook.Transaction, userId, accountId uuid.UUID, totalCount *int, limit, offset int) error {

	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	accountOwner := user.User{}
	if err := service.repository.GetRecordByID(uow, userId, &accountOwner); err != nil {
		return errors.NewDatabaseError("user not found")
	}
	if accountOwner.IsActive != nil && !*accountOwner.IsActive {
		return errors.NewInActiveUserError("can not get the passbook records for InActive user")
	}

	userAccount := account.Account{}
	if err := service.repository.GetRecordByID(uow, accountId, &userAccount); err != nil {
		return errors.NewNotFoundError("Unable fetch account details")
	}
	if userAccount.IsActive != nil && !*userAccount.IsActive {
		return errors.NewInActiveUserError("can not get the passbook records for InActive Account")
	}

	accountOwnerBank := bank.Bank{}
	if err := service.repository.GetRecordByID(uow, userAccount.BankID, &accountOwnerBank); err != nil {
		return errors.NewValidationError("bank not found")
	}
	if accountOwnerBank.IsActive != nil && !*accountOwnerBank.IsActive {
		return errors.NewInActiveUserError("can not get the passbook records for InActive Bank")
	}

	queryProcessor := []repository.QueryProcessor{
		repository.Filter("account_id = ?", userAccount.ID),
		repository.Paginate(limit, offset, totalCount),
	}
	err := service.repository.GetAll(uow, passbook, queryProcessor...)
	if err != nil {
		return err
	}

	err = service.repository.GetCount(uow, passbook, totalCount, repository.Filter("account_id = ?", userAccount.ID))
	if err != nil {
		return err
	}

	return nil
}
