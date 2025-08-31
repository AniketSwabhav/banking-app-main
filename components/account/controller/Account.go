package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/account"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	accountService "banking-app-be/components/account/service"
)

type AccountController struct {
	log            log.Logger
	AccountService *accountService.AccountService
}

func NewAccountController(accountService *accountService.AccountService, log log.Logger) *AccountController {
	return &AccountController{
		log:            log,
		AccountService: accountService,
	}
}

func (Controller *AccountController) RegisterRoutes(router *mux.Router) {

	// http://localhost:8001/api/v1/banking-app/
	accountRouter := router.PathPrefix("/account").Subrouter()
	guardedRouter := accountRouter.PathPrefix("/").Subrouter()

	//Post
	guardedRouter.HandleFunc("/bank/{bankId}", Controller.createAccount).Methods(http.MethodPost)

	//Get
	guardedRouter.HandleFunc("/", Controller.getAllUserAccounts).Methods(http.MethodGet)
	guardedRouter.HandleFunc("/{id}", Controller.getAccountByAccountID).Methods(http.MethodGet)

	//Withdraw
	guardedRouter.HandleFunc("/withdraw", Controller.withdrawFromAccount).Methods(http.MethodPost)

	//Deposite
	guardedRouter.HandleFunc("/deposite", Controller.depositetToAccount).Methods(http.MethodPost)

	//Delete
	guardedRouter.HandleFunc("/{id}", Controller.deleteAccountByAccountID).Methods(http.MethodDelete)
	guardedRouter.Use(security.MiddlewareUser)

	//Transfer
	guardedRouter.HandleFunc("/transfer", Controller.transfer).Methods(http.MethodPost)

	guardedRouter.Use(security.MiddlewareUser)
}

func (controller *AccountController) createAccount(w http.ResponseWriter, r *http.Request) {

	newAccount := account.Account{}
	parser := web.NewParser(r)

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}
	newAccount.UserID = userID
	newAccount.CreatedBy = userID

	bankID, err := parser.GetUUID("bankId")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid bank ID format"))
		return
	}
	newAccount.BankID = bankID

	err = controller.AccountService.CreateAccount(&newAccount)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newAccount)
}

func (controller *AccountController) getAllUserAccounts(w http.ResponseWriter, r *http.Request) {

	allAccounts := []account.AccountDTO{}
	var totalCount int
	query := r.URL.Query()

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5 //default
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 //default
	}

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}

	err = controller.AccountService.GetAllAccountsByUserID(userID, &allAccounts, &totalCount, limit, offset)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allAccounts)
}

func (controller *AccountController) getAccountByAccountID(w http.ResponseWriter, r *http.Request) {

	accountToGet := account.AccountDTO{}

	parser := web.NewParser(r)

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}

	accountIDFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid Account ID format"))
		return
	}

	accountToGet.UserID = userID
	accountToGet.ID = accountIDFromURL

	err = controller.AccountService.GetAccountByAccountID(&accountToGet)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, accountToGet)

}

func (controller *AccountController) deleteAccountByAccountID(w http.ResponseWriter, r *http.Request) {

	accountToDelete := account.Account{}
	parser := web.NewParser(r)

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}

	accountIDFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid Account ID format"))
		return
	}

	accountToDelete.UserID = userID
	accountToDelete.ID = accountIDFromURL
	accountToDelete.DeletedBy = userID

	err = controller.AccountService.DeleteAccountById(&accountToDelete)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})

}

func (controller *AccountController) withdrawFromAccount(w http.ResponseWriter, r *http.Request) {

	accountToUpdate := account.Account{}

	var requestData struct {
		AccountNo string  `json:"accountNo"`
		Amount    float32 `json:"amount"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	accountToUpdate.AccountNo = requestData.AccountNo

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}
	accountToUpdate.UserID = userID
	accountToUpdate.UpdatedBy = userID

	err = controller.AccountService.Withdraw(accountToUpdate, requestData.Amount)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Withdrawal successful",
		// "account_no":      accountToUpdate.AccountNo,
		// "updated_balance": accountToUpdate.AccountBalance,
	})
}

func (controller *AccountController) depositetToAccount(w http.ResponseWriter, r *http.Request) {

	accountToUpdate := account.Account{}

	var requestData struct {
		AccountNo string  `json:"accountNo"`
		Amount    float32 `json:"amount"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	accountToUpdate.AccountNo = requestData.AccountNo

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}
	accountToUpdate.UserID = userID
	accountToUpdate.UpdatedBy = userID

	err = controller.AccountService.Deposite(accountToUpdate, requestData.Amount)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Money Deposited successful",
		// "account_no":      accountToUpdate.AccountNo,
		// "updated_balance": accountToUpdate.AccountBalance,
	})
}

func (controller *AccountController) transfer(w http.ResponseWriter, r *http.Request) {

	fromAccount := account.Account{}
	toAccount := account.Account{}

	var requestData struct {
		FromAccountNo string  `json:"fromAccountNo"`
		ToAccountNo   string  `json:"toAccountNo"`
		Amount        float32 `json:"amount"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	fromAccount.AccountNo = requestData.FromAccountNo
	fmt.Println("from account number =======================>", fromAccount.AccountNo)
	toAccount.AccountNo = requestData.ToAccountNo

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}

	fromAccount.UserID = userID
	fromAccount.UpdatedBy = userID
	toAccount.UpdatedBy = userID

	err = controller.AccountService.Transfer(fromAccount, toAccount, requestData.Amount)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Money Transferred successfully",
	})

}
