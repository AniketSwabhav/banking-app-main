package service

import (
	"banking-app-be/components/errors"
	"banking-app-be/model/account"
	"banking-app-be/model/bank"
	"banking-app-be/model/passbook"
	"banking-app-be/model/user"
	"banking-app-be/module/repository"
	"fmt"
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

	accountOwner := user.User{}
	if err := service.repository.GetRecordByID(uow, newAccount.UserID, &accountOwner); err != nil {
		return errors.NewNotFoundError("Account owner not found")
	}

	if !accountOwner.IsActive {
		return errors.NewInActiveUserError("Can not create a Bank account for InActive user")
	}

	bank := bank.Bank{}
	if err := service.repository.GetRecordByID(uow, newAccount.BankID, &bank); err != nil {
		return errors.NewNotFoundError("Bank not found with given Id")
	}

	if !accountOwner.IsActive {
		return errors.NewInActiveUserError("Can not create a account in InActive bank")
	}

	accountNo, err := service.generateUniqueAccountNumber()
	if err != nil {
		return err
	}
	newAccount.AccountNo = accountNo

	err = newAccount.Validate()
	if err != nil {
		return err
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

	if err := service.repository.Add(uow, newAccount); err != nil {
		return errors.NewDatabaseError("Failed to create account")
	}

	accountOwner.TotalBalance += newAccount.AccountBalance

	updateOwner := map[string]interface{}{
		"total_balance": accountOwner.TotalBalance,
	}

	if err := service.repository.UpdateWithMap(uow, &user.User{}, updateOwner, repository.Filter("id = ?", newAccount.UserID)); err != nil {
		return errors.NewDatabaseError("Failed to update Owners total balance")
	}

	fmt.Println(accountOwner)

	uow.Commit()
	return nil
}

func (service *AccountService) GetAllAccountsByUserID(userID uuid.UUID, allAccounts *[]account.AccountDTO, totalCount *int, limit, offset int) error {
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

func (service *AccountService) GetAccountByAccountID(accountToGet *account.AccountDTO) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	if err := service.repository.GetRecordByID(uow, accountToGet.ID, accountToGet); err != nil {
		return errors.NewNotFoundError("Account not found with given Id")
	}

	tempAccount := account.Account{}
	if err := service.repository.GetRecord(uow, tempAccount, repository.Filter("id = ? AND user_id = ?", accountToGet.ID, accountToGet.UserID)); err != nil {
		return errors.NewNotFoundError("does not match accountId and UserId")
	}

	if tempAccount.UserID != accountToGet.UserID {
		return errors.NewUnauthorizedError("AccountId is not related to Current User")
	}

	uow.Commit()
	return nil
}

func (service *AccountService) DeleteAccountById(accountToDelete *account.Account) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	if err := service.repository.GetRecordByID(uow, accountToDelete.ID, accountToDelete); err != nil {
		return errors.NewNotFoundError("Account not found with given Id")
	}

	tempAccount := account.Account{}
	if err := service.repository.GetRecord(uow, &tempAccount, repository.Filter("id = ? AND user_id = ?", accountToDelete.ID, accountToDelete.UserID)); err != nil {
		return errors.NewHTTPError("Account not found with given Account Number for Current User ", http.StatusNotFound)
	}

	if tempAccount.UserID != accountToDelete.UserID {
		return errors.NewUnauthorizedError("AccountId is not related to Current User")
	}

	updateData := map[string]interface{}{
		"is_active":  false,
		"deleted_by": accountToDelete.DeletedBy,
		"deleted_at": time.Now(),
	}
	if err := service.repository.UpdateWithMap(uow, &account.Account{}, updateData, repository.Filter("id = ? AND user_id = ?", accountToDelete.ID, accountToDelete.UserID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (service *AccountService) Withdraw(accountToUpdate account.Account, amount float32) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	accountOwner := user.User{}
	if err := service.repository.GetRecordByID(uow, accountToUpdate.UserID, &accountOwner); err != nil {
		return errors.NewNotFoundError("Account owner not found")
	}
	if !accountOwner.IsActive {
		return errors.NewInActiveUserError("InActive user can not withdraw money")
	}

	existingAccount := account.Account{}
	if err := service.repository.GetRecord(uow, &existingAccount, repository.Filter("account_no = ? AND user_id = ?", accountToUpdate.AccountNo, accountToUpdate.UserID)); err != nil {
		return errors.NewHTTPError("Account not found with given Account Number for Current User ", http.StatusNotFound)
	}

	bank := bank.Bank{}
	if err := service.repository.GetRecordByID(uow, existingAccount.BankID, &bank); err != nil {
		return errors.NewNotFoundError("invalid bankid")
	}
	if !bank.IsActive {
		return errors.NewInActiveUserError("Can not withdraw money from InActive Bank")
	}

	if existingAccount.AccountBalance < float32(amount) {
		return errors.NewValidationError("Insufficient balance")
	}

	accountToUpdate.AccountBalance = existingAccount.AccountBalance - float32(amount)

	transaction := passbook.Transaction{
		TimeStamp:      time.Now(),
		Type:           "Withdrawal",
		Amount:         -float32(amount),
		AccountBalance: accountToUpdate.AccountBalance,
		Note:           "Withdrawal transaction",
		AccountID:      existingAccount.ID,
	}

	if err := uow.DB.Create(&transaction).Error; err != nil {
		return errors.NewDatabaseError("Failed to record transaction")
	}

	updateData := map[string]interface{}{
		"account_balance": accountToUpdate.AccountBalance,
		"updated_by":      accountToUpdate.UpdatedBy,
		"updated_at":      time.Now(),
	}
	if err := service.repository.UpdateWithMap(uow, &account.Account{}, updateData, repository.Filter("account_no = ? AND user_id = ?", accountToUpdate.AccountNo, accountToUpdate.UserID)); err != nil {
		return err
	}

	accountOwner.TotalBalance -= amount

	updateOwnerData := map[string]interface{}{
		"total_balance": accountOwner.TotalBalance,
		"updated_by":    accountToUpdate.UserID,
		"updated_at":    time.Now(),
	}
	if err := service.repository.UpdateWithMap(uow, &user.User{}, updateOwnerData, repository.Filter("id = ?", accountToUpdate.UserID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (service *AccountService) Deposite(accountToUpdate account.Account, amount float32) error {

	uow := repository.NewUnitOfWork(service.db, false)
	defer uow.RollBack()

	accountOwner := user.User{}
	if err := service.repository.GetRecordByID(uow, accountToUpdate.UserID, &accountOwner); err != nil {
		return errors.NewNotFoundError("Account owner not found")
	}
	if !accountOwner.IsActive {
		return errors.NewInActiveUserError("InActive user can not withdraw money")
	}

	existingAccount := account.Account{}
	if err := service.repository.GetRecord(uow, &existingAccount, repository.Filter("account_no = ? AND user_id = ?", accountToUpdate.AccountNo, accountToUpdate.UserID)); err != nil {
		return errors.NewHTTPError("Account not found with given Account Number for Current User ", http.StatusNotFound)
	}

	bank := bank.Bank{}
	if err := service.repository.GetRecordByID(uow, existingAccount.BankID, &bank); err != nil {
		return errors.NewNotFoundError("invalid bankid")
	}
	if !bank.IsActive {
		return errors.NewInActiveUserError("Can not withdraw money from InActive Bank")
	}

	accountToUpdate.AccountBalance = existingAccount.AccountBalance + float32(amount)

	transaction := passbook.Transaction{
		TimeStamp:      time.Now(),
		Type:           "Deposite",
		Amount:         +float32(amount),
		AccountBalance: accountToUpdate.AccountBalance,
		Note:           "Deposite transaction",
		AccountID:      existingAccount.ID,
	}

	if err := uow.DB.Create(&transaction).Error; err != nil {
		return errors.NewDatabaseError("Failed to record transaction")
	}

	updateData := map[string]interface{}{
		"account_balance": accountToUpdate.AccountBalance,
		"updated_by":      accountToUpdate.UpdatedBy,
		"updated_at":      time.Now(),
	}

	if err := service.repository.UpdateWithMap(uow, &account.Account{}, updateData, repository.Filter("account_no = ? AND user_id = ?", accountToUpdate.AccountNo, accountToUpdate.UserID)); err != nil {
		return err
	}

	accountOwner.TotalBalance += amount

	updateOwnerData := map[string]interface{}{
		"total_balance": accountOwner.TotalBalance,
		"updated_by":    accountToUpdate.UserID,
		"updated_at":    time.Now(),
	}
	if err := service.repository.UpdateWithMap(uow, &user.User{}, updateOwnerData, repository.Filter("id = ?", accountToUpdate.UserID)); err != nil {
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
