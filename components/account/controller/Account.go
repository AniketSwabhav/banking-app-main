package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/account"
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

	//Update
	guardedRouter.HandleFunc("/{id}", Controller.updateAccountByAccountID).Methods(http.MethodPut)

	//Delete
	guardedRouter.HandleFunc("/{id}", Controller.deleteAccountByAccountID).Methods(http.MethodDelete)
	guardedRouter.Use(security.MiddlewareUser)

	//Withdraw
	guardedRouter.HandleFunc("/{id}/withdraw", Controller.withdrawFromAccount).Methods(http.MethodPost)

	//Deposite
	guardedRouter.HandleFunc("/{id}/deposite", Controller.depositetToAccount).Methods(http.MethodPost)

	//Transfer
	guardedRouter.HandleFunc("/{id}/transfer", Controller.transfer).Methods(http.MethodPost)

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

	allAccounts := []account.AccontBankDTO{}
	var totalCount int
	query := r.URL.Query()

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
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

func (controller *AccountController) updateAccountByAccountID(w http.ResponseWriter, r *http.Request) {
	accountToUpdate := account.Account{}
	parser := web.NewParser(r)

	var err error

	accountToUpdate.UpdatedBy, err = security.ExtractUserIDFromToken(r)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}

	accountToUpdate.ID, err = parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid Account ID format"))
		return
	}
	accountToUpdate.UserID = accountToUpdate.UpdatedBy

	err = web.UnmarshalJSON(r, &accountToUpdate)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse requested data", http.StatusBadRequest))
		return
	}

	err = controller.AccountService.UpdateAccountById(&accountToUpdate)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, accountToUpdate)

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
	parser := web.NewParser(r)

	var requestData struct {
		// AccountNo string  `json:"accountNo"`
		Amount float32 `json:"amount"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	accountIDFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid Account ID format"))
		return
	}
	accountToUpdate.ID = accountIDFromURL

	// accountToUpdate.AccountNo = requestData.AccountNo

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
	})
}

func (controller *AccountController) depositetToAccount(w http.ResponseWriter, r *http.Request) {

	accountToUpdate := account.Account{}
	parser := web.NewParser(r)

	var requestData struct {
		// AccountNo string  `json:"accountNo"`
		Amount float32 `json:"amount"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	accountIDFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid Account ID format"))
		return
	}
	accountToUpdate.ID = accountIDFromURL

	// accountToUpdate.AccountNo = requestData.AccountNo

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
	})
}

func (controller *AccountController) transfer(w http.ResponseWriter, r *http.Request) {

	fromAccount := account.Account{}
	toAccount := account.Account{}
	parser := web.NewParser(r)

	var requestData struct {
		// FromAccountNo string  `json:"fromAccountNo"`
		ToAccountNo string  `json:"toAccountNo"`
		Amount      float32 `json:"amount"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	accountIDFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid Account ID format"))
		return
	}

	// fromAccount.AccountNo = requestData.FromAccountNo
	// fmt.Println("from account number =======================>", fromAccount.AccountNo)
	fromAccount.ID = accountIDFromURL
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
