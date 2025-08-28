package service

import (
	"banking-app-be/components/errors"
	"banking-app-be/model/account"
	bankModel "banking-app-be/model/bank"
	"banking-app-be/model/passbook"
	userModel "banking-app-be/model/user"
	"banking-app-be/module/repository"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type AccountService struct {
	db         *gorm.DB
	repository repository.Repository
}

func NewAccountService(DB *gorm.DB, repo repository.Repository) *AccountService {
	return &AccountService{
		db:         DB,
		repository: repo,
	}
}

func (service *AccountService) CreateAccount(newAccount *account.Account) error {
	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	accountNo, err := service.generateUniqueAccountNumber()
	if err != nil {
		return err
	}
	newAccount.AccountNo = accountNo

	err = newAccount.Validate()
	if err != nil {
		return err
	}

	var user userModel.User
	if err := uow.DB.Where("id = ? AND is_active = true", newAccount.UserID).First(&user).Error; err != nil {
		return errors.NewHTTPError("User not found or inactive", http.StatusNotFound)
	}

	var bank bankModel.Bank
	if err := uow.DB.Where("id = ? AND is_active = true", newAccount.BankID).First(&bank).Error; err != nil {
		return errors.NewHTTPError("Bank not found or inactive", http.StatusNotFound)
	}

	if newAccount.AccountBalance == 0 {
		newAccount.AccountBalance = 1000
	}

	newAccount.PassBook = []passbook.Transaction{
		{
			TimeStamp:      time.Now(),
			Type:           "AccountCreation",
			Amount:         newAccount.AccountBalance,
			AccountBalance: newAccount.AccountBalance,
			Note:           "Account created with initial balance Rs.1000",
		},
	}

	if err := uow.DB.Create(newAccount).Error; err != nil {
		uow.RollBack()
		return errors.NewDatabaseError("Failed to create account")
	}

	uow.Commit()
	return nil
}

func (service *AccountService) GetAccountsByUser(userID uuid.UUID, allAccounts *[]account.AccountDTO, totalCount *int, limit, offset int) error {
	uow := repository.NewUnitOfWork(service.db, true)
	defer uow.RollBack()

	queryProcessor := []repository.QueryProcessor{
		repository.Filter("user_id = ?", userID),
		repository.PreloadAssociations([]string{"PassBook"}),
		repository.Paginate(limit, offset, totalCount),
	}

	err := service.repository.GetAll(uow, allAccounts, queryProcessor...)
	if err != nil {
		return err
	}

	err = service.repository.GetCount(uow, allAccounts, totalCount)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}

//===================================================================================================================

func (service *AccountService) generateUniqueAccountNumber() (string, error) {
	const maxAttempts = 5
	const accountLength = 12

	rand.Seed(time.Now().UnixNano())

	for attempt := 0; attempt < maxAttempts; attempt++ {
		accountNo := generateRandomDigits(accountLength)

		var count int
		err := service.db.Model(&account.Account{}).
			Where("account_no = ?", accountNo).
			Count(&count).Error

		if err != nil {
			return "", errors.NewDatabaseError("Failed to check account number uniqueness")
		}

		if count == 0 {
			return accountNo, nil
		}
	}

	return "", errors.NewHTTPError("Failed to generate unique account number after several attempts", http.StatusInternalServerError)
}

func generateRandomDigits(length int) string {
	var digits string
	for i := 0; i < length; i++ {
		digits += strconv.Itoa(rand.Intn(10))
	}
	return digits
}
